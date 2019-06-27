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

package route

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	istiolisters "knative.dev/pkg/client/listers/istio/v1alpha3"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles a Route object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	// listers index properties about resources
	routeLister          kflisters.RouteLister
	virtualServiceLister istiolisters.VirtualServiceLister
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

	original, err := r.routeLister.Routes(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Errorf("Route %q no longer exists\n", name)
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
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

	} else if _, uErr := r.updateStatus(toReconcile); uErr != nil {
		logger.Warnw("Failed to update Route status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Route .
func (r *Reconciler) ApplyChanges(ctx context.Context, route *v1alpha1.Route) error {
	route.Status.InitializeConditions()
	virtualServiceName := resources.VirtualServiceName(route.Spec.Hostname, route.Spec.Domain, route.Spec.Path)

	// Sync VirtualService
	{
		desired, err := resources.MakeVirtualService(route)
		if err != nil {
			return err
		}

		actual, err := r.virtualServiceLister.VirtualServices(desired.GetNamespace()).Get(desired.Name)
		if errors.IsNotFound(err) {
			// VirtualService doesn't exist, make one.
			actual, err = r.SharedClientSet.Networking().VirtualServices(desired.GetNamespace()).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, route) {
			route.Status.MarkVirtualServiceNotOwned(virtualServiceName)
			return fmt.Errorf("route: %q does not own VirtualService: %q", route.Name, virtualServiceName)
		} else if actual, err = r.reconcile(desired, actual); err != nil {
			return err
		}

		route.Status.PropagateVirtualServiceStatus(actual)
	}

	return nil
}

func (r *Reconciler) reconcile(desired, actual *networking.VirtualService) (*networking.VirtualService, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff VirtualService: %v", err)
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.Spec = desired.Spec
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	return r.SharedClientSet.Networking().VirtualServices(existing.GetNamespace()).Create(existing)
}

func (r *Reconciler) updateStatus(desired *v1alpha1.Route) (*v1alpha1.Route, error) {
	actual, err := r.routeLister.Routes(desired.GetNamespace()).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().Routes(existing.GetNamespace()).UpdateStatus(existing)
}
