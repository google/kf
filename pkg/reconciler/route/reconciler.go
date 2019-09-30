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
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/reconciler"
	appresources "github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	routeClaimLister     kflisters.RouteClaimLister
	virtualServiceLister istiolisters.VirtualServiceLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by Kubernetes.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Key is a JSON marshalled namespacedRouteSpecFields
	var route namespacedRouteSpecFields
	if err := json.Unmarshal([]byte(key), &route); err != nil {
		return err
	}

	logger := logging.FromContext(ctx).With("namespace", route.Namespace)

	if r.IsNamespaceTerminating(route.Namespace) {
		logger.Errorf("skipping sync for route %#v", route)
		return nil
	}

	return r.ApplyChanges(
		logging.WithLogger(ctx, logger),
		route.Namespace,
		route.RouteSpecFields,
	)
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Route .
func (r *Reconciler) ApplyChanges(
	ctx context.Context,
	namespace string,
	fields v1alpha1.RouteSpecFields,
) error {
	logger := logging.FromContext(ctx)
	fields.SetDefaults(ctx)

	// Sync VirtualService
	logger.Debug("reconciling VirtualService")

	// Fetch Claims with the same Hostname+Domain+Path.
	claims, err := r.routeClaimLister.
		RouteClaims(namespace).
		List(appresources.MakeRouteSelector(fields))
	if err != nil {
		return err
	}

	// There aren't any claims, so there shouldn't be any VirtualServices
	// or Routes.
	// NOTE: The generated LabelSelectors have the ManagedBy kf
	// requirement. Therefore if an operator manually creates a route or
	// virtualservice, it won't be cleaned up.
	if len(claims) == 0 {
		err := r.SharedClientSet.
			Networking().
			VirtualServices(v1alpha1.KfNamespace).
			Delete(v1alpha1.GenerateName(
				fields.Hostname,
				fields.Domain,
			), &metav1.DeleteOptions{})

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		err = r.KfClientSet.
			Kf().
			Routes(namespace).
			DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
				LabelSelector: appresources.MakeRouteSelectorNoPath(fields).String(),
			})

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		return nil
	}

	// Fetch routes with the same Hostname+Domain+Path.
	routes, err := r.routeLister.
		Routes(namespace).
		List(appresources.MakeRouteSelector(fields))
	if err != nil {
		return err
	}

	desired, err := resources.MakeVirtualService(claims, routes)
	if err != nil {
		return err
	}

	actual, err := r.virtualServiceLister.
		VirtualServices(v1alpha1.KfNamespace).
		Get(desired.Name)
	if errors.IsNotFound(err) {
		// VirtualService doesn't exist, make one.
		if _, err := r.SharedClientSet.
			Networking().
			VirtualServices(v1alpha1.KfNamespace).
			Create(desired); err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	} else if actual.GetDeletionTimestamp() != nil {
		return nil
	} else if actual, err = r.update(
		ctx,
		desired,
		actual,
	); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) update(
	ctx context.Context,
	desired *networking.VirtualService,
	actual *networking.VirtualService,
) (*networking.VirtualService, error) {
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

	// Merge new OwnerReferences and HTTPRoutes
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

	return r.SharedClientSet.
		Networking().
		VirtualServices(existing.GetNamespace()).
		Update(existing)
}
