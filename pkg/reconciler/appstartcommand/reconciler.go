// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appstartcommand

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/go-containerregistry/pkg/name"
	containerregistryv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/dockerutil"
	"github.com/google/kf/v2/pkg/reconciler"
	appreconciler "github.com/google/kf/v2/pkg/reconciler/app"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

type Reconciler struct {
	*reconciler.Base
	appLister     kflisters.AppLister
	kfConfigStore *kfconfig.Store
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of the
// watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	return r.reconcileApp(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileApp(ctx context.Context, namespace, name string) (err error) {
	logger := logging.FromContext(ctx)
	original, err := r.appLister.Apps(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Debug("resource no longer exists")
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		logger.Debug("resource deletion requested")
		return nil
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Debug("namespace is terminating, skipping reconciliation")
		return nil
	}

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// ALWAYS update the ObservedGenration: "If the primary resource your
	// controller is reconciling supports ObservedGeneration in its status, make
	// sure you correctly set it to metadata.Generation whenever the values
	// between the two fields mismatches."
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
	toReconcile.Status.ObservedGeneration = toReconcile.Generation

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if reconcileErr != nil {
		logger.Debugf("App reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStartCommandStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("Failed to update App start command status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

func (r *Reconciler) ApplyChanges(ctx context.Context, app *v1alpha1.App) error {
	logger := logging.FromContext(ctx)
	ctx = r.kfConfigStore.ToContext(ctx)

	// Default values on the app in case it hasn't been triggered since last update
	// to spec.
	app.SetDefaults(ctx)

	// Sync start commands, populate container and buildpack start commands in app status.
	{
		configDefaults, err := kfconfig.FromContext(ctx).Defaults()
		if err != nil {
			return fmt.Errorf("failed to read config-defaults: %v", err)
		}

		if !configDefaults.AppDisableStartCommandLookup {
			logger.Debug("reconciling start commands")
			r.updateStartCommand(app, fetchContainerCommand)
		}
	}
	return nil
}

type ImageConfigFetcher func(image string) (*containerregistryv1.ConfigFile, error)

func (*Reconciler) updateStartCommand(app *v1alpha1.App, fetcher ImageConfigFetcher) {
	if app.Status.Image == app.Status.StartCommands.Image ||
		app.Status.Image == appreconciler.DefaultPlaceHolderBuildImage ||
		app.Status.Image == "" {
		// Don't lookup start commands if we've already cached them or expect them not to exist.
		return
	}

	startCommands := v1alpha1.StartCommandStatus{
		// The image is set to prevent repeatedly looking up the values for the same image.
		Image: app.Status.Image,
	}

	containerConfig, err := fetcher(app.Status.Image)
	if err != nil {
		startCommands.Error = err.Error()
	} else {
		startCommands.Container = containerConfig.Config.Entrypoint

		// Look for a special label set by the buildpack that might contain the
		// start command for v2 buildpacks.
		if maybeStartCommand, ok := containerConfig.Config.Labels["StartCommand"]; ok {
			startCommands.Buildpack = []string{maybeStartCommand}
		}
	}

	app.Status.PropagateStartCommandStatus(startCommands)
}

func fetchContainerCommand(image string) (*containerregistryv1.ConfigFile, error) {
	imageRef, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(imageRef, dockerutil.GetAuthKeyChain())
	if err != nil {
		return nil, err
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return configFile, nil
}

func (r *Reconciler) updateStartCommandStatus(ctx context.Context, desired *v1alpha1.App) (*v1alpha1.App, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating status")
	actual, err := r.appLister.Apps(desired.GetNamespace()).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status.StartCommands, desired.Status.StartCommands) {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()
	existing.Status.StartCommands = desired.Status.StartCommands

	return r.KfClientSet.KfV1alpha1().Apps(existing.GetNamespace()).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}
