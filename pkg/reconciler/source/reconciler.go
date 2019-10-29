// Copyright 2019 Google LLC
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

package source

import (
	"context"
	"reflect"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/source/resources"
	buildclient "github.com/google/kf/third_party/knative-build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	buildlisters "github.com/google/kf/third_party/knative-build/pkg/client/listers/build/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles an source object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	buildClient buildclient.BuildV1alpha1Interface

	// listers index properties about resources
	sourceLister kflisters.SourceLister
	buildLister  buildlisters.BuildLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by Kubernetes.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	return r.reconcileSource(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileSource(
	ctx context.Context,
	namespace string,
	name string,
) (err error) {
	logger := logging.FromContext(ctx)

	original, err := r.sourceLister.Sources(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		logger.Errorf("source %q no longer exists\n", name)
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		return nil
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Errorf("skipping sync for source %q, namespace %q is terminating\n", name, namespace)
		return nil
	}

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(namespace, toReconcile); uErr != nil {
		logger.Warnw("Failed to update Source status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the source.
func (r *Reconciler) ApplyChanges(ctx context.Context, source *v1alpha1.Source) error {
	logger := logging.FromContext(ctx)
	source.Status.InitializeConditions()

	// Sources only get run once regardless of success or failure status.
	if v1alpha1.IsStatusFinal(source.Status.Status) {
		return nil
	}

	secretCondition := source.Status.BuildSecretCondition()
	buildCondtion := source.Status.BuildCondition()

	desiredBuild, desiredSecret, err := resources.MakeBuild(source)
	if err != nil {
		return secretCondition.MarkTemplateError(err)
	}

	// Sync Build Secret
	if desiredSecret != nil {
		logger.Debug("reconciling build secret")

		actual, err := r.SecretLister.
			Secrets(desiredSecret.Namespace).
			Get(desiredSecret.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				CoreV1().
				Secrets(desiredSecret.Namespace).
				Create(desiredSecret)
			if err != nil {
				return secretCondition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return secretCondition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, source) {
			return secretCondition.MarkChildNotOwned(desiredSecret.Name)
		} else if actual, err = r.ReconcileSecret(ctx, desiredSecret, actual); err != nil {
			return secretCondition.MarkReconciliationError("updating existing", err)
		}
		source.Status.PropagateBuildSecretStatus(actual)

		if secretCondition.IsPending() {
			logger.Info("Waiting for Secret; exiting early")
			return nil
		}
	}

	// Sync Build
	{
		logger.Debug("reconciling Build")

		actual, err := r.buildLister.
			Builds(source.Namespace).
			Get(desiredBuild.Name)
		if errors.IsNotFound(err) {
			actual, err = r.buildClient.
				Builds(desiredBuild.Namespace).
				Create(desiredBuild)
			if err != nil {
				return err
			}
		} else if !metav1.IsControlledBy(actual, source) {
			return buildCondtion.MarkChildNotOwned(desiredBuild.Name)
		}

		source.Status.PropagateBuildStatus(actual)
	}

	return nil
}

func (r *Reconciler) updateStatus(namespace string, desired *v1alpha1.Source) (*v1alpha1.Source, error) {
	actual, err := r.sourceLister.Sources(namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}

	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()
	existing.Status = desired.Status

	return r.KfClientSet.KfV1alpha1().Sources(namespace).UpdateStatus(existing)
}
