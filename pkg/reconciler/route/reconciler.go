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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	networkingclientset "github.com/google/kf/v2/pkg/client/networking/clientset/versioned"
	networkinglisters "github.com/google/kf/v2/pkg/client/networking/listers/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/route/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles a Route object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	// listers index properties about resources
	routeLister                  kflisters.RouteLister
	virtualServiceLister         networkinglisters.VirtualServiceLister
	networkingClientSet          networkingclientset.Interface
	appLister                    kflisters.AppLister
	spaceLister                  kflisters.SpaceLister
	serviceInstanceBindingLister kflisters.ServiceInstanceBindingLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of the
// watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, domain, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	logger := logging.FromContext(ctx).With(
		zap.Reflect("namespace", namespace),
		zap.Reflect("domain", domain),
	)

	if r.IsNamespaceTerminating(namespace) {
		logger.Errorf("skipping sync for domain %#v", domain)
		return nil
	}

	return r.ApplyChanges(
		logging.WithLogger(ctx, logger),
		namespace,
		domain,
	)
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Route.
func (r *Reconciler) ApplyChanges(
	ctx context.Context,
	namespace string,
	domain string,
) error {
	logger := logging.FromContext(ctx)

	// Check that the domain is valid for the Space
	var spaceDomain *v1alpha1.SpaceDomain
	if space, err := r.spaceLister.Get(namespace); err != nil {
		logger.Warnw("Failed to get Space", zap.Error(err))
		return err
	} else {
		for _, sd := range space.Status.NetworkConfig.Domains {
			if sd.Domain == domain {
				spaceDomain = sd.DeepCopy()
				break
			}
		}
	}
	logger = logger.With(zap.Reflect("spaceDomain", spaceDomain))

	// Sync VirtualService
	logger.Debug("reconciling VirtualService")

	// Fetch Routes with the same Domain.
	routesOrig, err := r.routeLister.
		Routes(namespace).
		List(labels.Everything())
	if err != nil {
		return err
	}

	routes := []*v1alpha1.Route{}
	for _, r := range routesOrig {
		tmp := r.DeepCopy()
		tmp.SetDefaults(ctx)
		if tmp.Spec.Domain == domain {
			routes = append(routes, tmp)
		}
	}
	logger = logger.With(zap.Reflect("routes", routes))

	// Fetch Apps that are bound to the Routes.
	apps, err := r.appLister.
		Apps(namespace).
		List(labels.Everything())
	if err != nil {
		return err
	}

	// appBindings is a map of RouteSpecFields strings to bound App destinations.
	// Keys are strings to normalize differences between equal RouteSpecFields (e.g. Path being "" or "/")
	appBindings := make(map[string]resources.RouteBindingSlice)
	for _, app := range apps {
		a := app.DeepCopy()
		a.SetDefaults(ctx)

		for _, binding := range a.Status.Routes {
			// Add all routes that have the same domain as the one being reconciled.
			// Don't reconcile bindings that are orphaned to prevent infinite loops.
			if binding.Source.Domain == domain && binding.Status != v1alpha1.RouteBindingStatusOrphaned {
				rsfString := binding.Source.String()
				appBindings[rsfString] = append(appBindings[rsfString], binding.Destination)
			}
		}
	}
	logger = logger.With(zap.Reflect("appBindings", appBindings))

	// Sort destinations in appBindings so the order is deterministic
	for rsf, appDestinations := range appBindings {
		sort.Sort(appDestinations)
		appBindings[rsf] = appDestinations
	}

	// Fetch route services that are bound to the Routes with the same domain.
	serviceBindings, err := r.serviceInstanceBindingLister.
		ServiceInstanceBindings(namespace).
		List(labels.Everything())
	if err != nil {
		return err
	}

	// routeServiceBindings is a map of RouteSpecFields strings to bound route service destinations.
	// A Route should have at most one route service bound to it.
	// However, the value in this map is a list of route service destinations, in the unlikely case that more than one route service is bound.
	// The RouteServiceReady condition for the Route is updated to False in this case.
	routeServiceBindings := make(map[string][]v1alpha1.RouteServiceDestination)

	for _, binding := range serviceBindings {
		routeRef := binding.Spec.BindingType.Route
		if routeRef != nil && routeRef.Domain == domain {
			rsfString := v1alpha1.RouteSpecFields(*routeRef).String()
			routeServiceDestination := v1alpha1.RouteServiceDestination{
				Name:            binding.Spec.InstanceRef.Name,
				RouteServiceURL: binding.Status.RouteServiceURL,
			}
			routeServiceBindings[rsfString] = append(routeServiceBindings[rsfString], routeServiceDestination)
		}
	}
	logger = logger.With(zap.Reflect("routeServiceBindings", routeServiceBindings))

	// Create or update VirtualService
	actualVS, sErr := r.reconcileVirtualService(
		logging.WithLogger(ctx, logger),
		namespace,
		domain,
		routes,
		appBindings,
		routeServiceBindings,
		spaceDomain,
	)

	// Used if the reconciler should fail based on the conditions of the Routes
	var exitErr error

	// Update Route statuses
	for _, origRoute := range routes {
		// Don't modify the informers copy
		toReconcile := origRoute.DeepCopy()

		// ALWAYS update the ObservedGenration: "If the primary resource your
		// controller is reconciling supports ObservedGeneration in its
		// status, make sure you correctly set it to metadata.Generation
		// whenever the values between the two fields mismatches."
		// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
		toReconcile.Status.ObservedGeneration = toReconcile.Generation

		toReconcile.Status.PropagateVirtualService(actualVS, sErr)
		toReconcile.Status.PropagateRouteSpecFields(origRoute.Spec.RouteSpecFields)

		rsfString := toReconcile.Spec.RouteSpecFields.String()
		toReconcile.Status.PropagateBindings(appBindings[rsfString])
		toReconcile.Status.PropagateRouteServiceBinding(routeServiceBindings[rsfString])
		toReconcile.Status.PropagateSpaceDomain(spaceDomain)

		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the
		// informer's cache may be stale and we don't want to overwrite a
		// prior update to status with this stale state.
		if !equality.Semantic.DeepEqual(origRoute.Status, toReconcile.Status) {
			// Update status
			if _, err := r.KfClientSet.
				KfV1alpha1().
				Routes(origRoute.GetNamespace()).
				UpdateStatus(ctx, toReconcile, metav1.UpdateOptions{}); err != nil {

				// Failed to update status
				logger.Warnw("Failed to update Route status", zap.Error(err))
				if exitErr != nil {
					exitErr = err
				}
			}
		}
	}

	// Prioritize error outputs so the root cause bubbles up
	if sErr != nil {
		return fmt.Errorf("Error occurred while reconciling VirtualService: %s", sErr.Error())
	}
	return exitErr
}

func (r *Reconciler) reconcileVirtualService(
	ctx context.Context,
	namespace string,
	domain string,
	routes []*v1alpha1.Route,
	appBindings map[string]resources.RouteBindingSlice,
	routeServiceBindings map[string][]v1alpha1.RouteServiceDestination,
	spaceDomain *v1alpha1.SpaceDomain,
) (*networking.VirtualService, error) {
	logger := logging.FromContext(ctx)

	// If the domain isn't permitted or there are no routes, the VS should be
	// removed. It will be removed anyway if there are no routes by the Kubernetes
	// GC, however we'll do it anyway for consistency.
	if spaceDomain == nil || len(routes) == 0 {
		logger.Warn("Deleting VirtualService because Space doesn't permit domain or there are no Routes")

		vsName := resources.MakeVirtualServiceName(domain)
		err := r.networkingClientSet.
			NetworkingV1alpha3().
			VirtualServices(namespace).
			Delete(ctx, vsName, metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			return nil, nil
		}
		// Pass error back exactly so API status code is preserved
		return nil, err
	}

	desired, err := resources.MakeVirtualService(routes, appBindings, routeServiceBindings, spaceDomain)
	if err != nil {
		return nil, fmt.Errorf("configuring: %v", err)
	}
	logger = logger.With(zap.Reflect("desired", desired))

	actual, err := r.virtualServiceLister.
		VirtualServices(namespace).
		Get(desired.Name)
	if errors.IsNotFound(err) {
		// VirtualService doesn't exist, make one.
		if actual, err = r.networkingClientSet.
			NetworkingV1alpha3().
			VirtualServices(namespace).
			Create(ctx, desired, metav1.CreateOptions{}); err != nil {
			logger.Errorf("creating: %v", err)
			// Pass error back exactly so API status code is preserved
			return nil, err
		}
	} else if err != nil {
		// Pass error back exactly so API status code is preserved
		return nil, err
	} else if actual.GetDeletionTimestamp() != nil {
		return nil, nil
	} else if actual, err = r.update(
		ctx,
		desired,
		actual,
	); err != nil {
		logger.Errorf("updating: %v", err)
		// Pass error back exactly so API status code is preserved
		return nil, err
	}

	return actual, nil
}

func (r *Reconciler) update(
	ctx context.Context,
	desired *networking.VirtualService,
	actual *networking.VirtualService,
) (*networking.VirtualService, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEquality := reconciler.NewSemanticEqualityBuilder(logger, "VirtualService").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("metadata.ownerReferences", desired.ObjectMeta.OwnerReferences, actual.ObjectMeta.OwnerReferences).
		Append("spec.gateways", desired.Spec.Gateways, actual.Spec.Gateways).
		Append("spec.hosts", desired.Spec.Hosts, actual.Spec.Hosts)

	if len(desired.Spec.Http) == len(actual.Spec.Http) {
		for i, http := range desired.Spec.Http {
			semanticEquality.Append(fmt.Sprintf("spec.http[%d]", i), http, actual.Spec.Http[i])
		}

		if semanticEquality.IsSemanticallyEqual() {
			return actual, nil
		}
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.OwnerReferences = desired.OwnerReferences

	// Set HTTP Routes, Hosts and Gateways
	existing.Spec.Http = desired.Spec.Http
	existing.Spec.Hosts = desired.Spec.Hosts
	existing.Spec.Gateways = desired.Spec.Gateways

	return r.networkingClientSet.
		NetworkingV1alpha3().
		VirtualServices(existing.GetNamespace()).
		Update(ctx, existing, metav1.UpdateOptions{})
}
