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
	"sort"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/route/resources"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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

	return r.reconcileRoute(ctx, namespace, name, logger)
}

func (r *Reconciler) reconcileRoute(ctx context.Context, namespace, name string, logger *zap.SugaredLogger) (err error) {
	var deleted bool
	original, err := r.routeLister.Routes(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Errorf("Route %q no longer exists\n", name)
		deleted = true

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		return nil
	}

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile, deleted, logger)
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

func (r *Reconciler) ReconcileServiceDeletion(ctx context.Context, service *serving.Service) error {
	// TODO(poy): This is O(n) where n is the number of apps in the
	// namespace. We can get this down to O(1).
	routes, err := r.routeLister.Routes(service.GetNamespace()).List(labels.Everything())
	if err != nil {
		return err
	}

	for _, route := range routes {
		for i, ksvcName := range route.Spec.KnativeServiceNames {
			if service.GetName() != ksvcName {
				continue
			}

			// Don't modify the informers copy
			toReconcile := route.DeepCopy()

			// Remove the Knative Service
			toReconcile.Spec.KnativeServiceNames = append(
				toReconcile.Spec.KnativeServiceNames[:i],
				toReconcile.Spec.KnativeServiceNames[i+1:]...,
			)

			// Update Route to not reference service
			if _, err := r.KfClientSet.
				KfV1alpha1().
				Routes(service.GetNamespace()).
				Update(toReconcile); err != nil {
				return err
			}
		}
	}

	return nil
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Route .
func (r *Reconciler) ApplyChanges(ctx context.Context, route *v1alpha1.Route, deleted bool, logger *zap.SugaredLogger) error {
	route.SetDefaults(ctx)
	route.Status.InitializeConditions()

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
		} else if actual, err = r.reconcile(desired, actual, deleted, logger); err != nil {
			return err
		}

		route.Status.PropagateVirtualServiceStatus(actual)
	}

	return nil
}

func (r *Reconciler) reconcile(desired, actual *networking.VirtualService, deleted bool, logger *zap.SugaredLogger) (*networking.VirtualService, error) {
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
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels

	if deleted {
		existing.OwnerReferences = algorithms.Delete(
			resources.OwnerReferences(existing.OwnerReferences),
			resources.OwnerReferences(desired.OwnerReferences),
		).(resources.OwnerReferences)

		existing.Spec.HTTP = algorithms.Delete(
			resources.HTTPRoutes(existing.Spec.HTTP),
			resources.HTTPRoutes(desired.Spec.HTTP),
		).(resources.HTTPRoutes)
	} else {
		existing.OwnerReferences = algorithms.Merge(
			resources.OwnerReferences(existing.OwnerReferences),
			resources.OwnerReferences(desired.OwnerReferences),
		).(resources.OwnerReferences)

		existing.Spec.HTTP = algorithms.Merge(
			resources.HTTPRoutes(existing.Spec.HTTP),
			resources.HTTPRoutes(desired.Spec.HTTP),
		).(resources.HTTPRoutes)

		// Sort by reverse to defer to the longest matchers.
		sort.Sort(sort.Reverse(resources.HTTPRoutes(existing.Spec.HTTP)))
	}

	return r.SharedClientSet.
		Networking().
		VirtualServices(existing.GetNamespace()).
		Update(existing)
}

func (r *Reconciler) updateStatus(desired *v1alpha1.Route) (*v1alpha1.Route, error) {
	actual, err := r.routeLister.
		Routes(desired.GetNamespace()).
		Get(desired.Name)
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

	return r.KfClientSet.
		KfV1alpha1().
		Routes(existing.GetNamespace()).
		UpdateStatus(existing)
}
