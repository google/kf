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
	"sort"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/reconciler"
	appresources "github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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

func (r *Reconciler) reconcileRoute(
	ctx context.Context,
	namespace string,
	name string,
	logger *zap.SugaredLogger,
) (err error) {
	var deleted bool
	original, err := r.routeLister.
		Routes(namespace).
		Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Errorf("Route %q no longer exists\n", name)
		deleted = true

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		return nil
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Errorf("skipping sync for route %q, namespace %q is terminating\n", name, namespace)
		return nil
	}

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// Reconcile this copy of the route.
	return r.ApplyChanges(ctx, toReconcile, deleted, logger)
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Route .
func (r *Reconciler) ApplyChanges(
	ctx context.Context,
	origRoute *v1alpha1.Route,
	deleted bool,
	logger *zap.SugaredLogger,
) error {
	origRoute.SetDefaults(ctx)

	// Sync VirtualService
	{
		// Fetch routes with the same Hostname+Domain+Path.
		routes, err := r.routeLister.
			Routes(origRoute.GetNamespace()).
			List(appresources.MakeRouteSelector(origRoute.Spec.RouteSpecFields))
		if err != nil {
			return err
		}

		desired, err := resources.MakeVirtualService(routes)
		if err != nil {
			return err
		}

		actual, err := r.virtualServiceLister.
			VirtualServices(desired.GetNamespace()).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			// VirtualService doesn't exist, make one.
			actual, err = r.SharedClientSet.
				Networking().
				VirtualServices(desired.GetNamespace()).
				Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if actual, err = r.reconcile(
			desired,
			actual,
			deleted,
			logger,
		); err != nil {
			return err
		}
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
	existing.ObjectMeta.Annotations = desired.ObjectMeta.Annotations

	if deleted {
		existing.OwnerReferences = algorithms.Delete(
			v1alpha1.OwnerReferences(existing.OwnerReferences),
			v1alpha1.OwnerReferences(desired.OwnerReferences),
		).(v1alpha1.OwnerReferences)

		existing.Spec.HTTP = algorithms.Delete(
			v1alpha1.HTTPRoutes(existing.Spec.HTTP),
			v1alpha1.HTTPRoutes(desired.Spec.HTTP),
		).(v1alpha1.HTTPRoutes)
	} else {
		existing.OwnerReferences = algorithms.Merge(
			v1alpha1.OwnerReferences(existing.OwnerReferences),
			v1alpha1.OwnerReferences(desired.OwnerReferences),
		).(v1alpha1.OwnerReferences)

		existing.Spec.HTTP = algorithms.Merge(
			v1alpha1.HTTPRoutes(existing.Spec.HTTP),
			v1alpha1.HTTPRoutes(desired.Spec.HTTP),
		).(v1alpha1.HTTPRoutes)

		// Sort by reverse to defer to the longest matchers.
		sort.Sort(sort.Reverse(v1alpha1.HTTPRoutes(existing.Spec.HTTP)))
	}

	return r.SharedClientSet.
		Networking().
		VirtualServices(existing.GetNamespace()).
		Update(existing)
}
