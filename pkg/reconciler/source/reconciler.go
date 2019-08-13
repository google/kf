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
	"fmt"
	"reflect"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	buildclient "github.com/google/kf/pkg/client/build/clientset/versioned/typed/build/v1alpha1"
	buildlisters "github.com/google/kf/pkg/client/build/listers/build/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/source/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
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
	logger := logging.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

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
	source.Status.InitializeConditions()

	// Sources only get run once regardless of success or failure status.
	if v1alpha1.IsStatusFinal(source.Status.Status) {
		return nil
	}

	// Sync build
	{
		desired, err := resources.MakeBuild(source)
		if err != nil {
			return err
		}

		actual, err := r.buildLister.Builds(source.Namespace).Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.buildClient.Builds(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if !metav1.IsControlledBy(actual, source) {
			source.Status.MarkBuildNotOwned(desired.Name)
			return fmt.Errorf("source: %q does not own build: %q", source.Name, desired.Name)
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
