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

package v1alpha1

import (
	"errors"
	"strings"

	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/kmeta"
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (r *Route) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Route")
}

var _ kmeta.OwnerRefable = (*Route)(nil)

// PropagateVirtualService stores the VirtualService in the RouteStatus.
// If vsErr is set, a reconciliation error is triggered and the name will be
// empty. The state is unknown if both VirtualService and error are nil.
// If shouldTrack is false, directly setting the condition to True.
func (status *RouteStatus) PropagateVirtualService(vs *networking.VirtualService, vsErr error, shouldTrack bool) {
	cond := status.VirtualServiceCondition()

	switch {
	case !shouldTrack:
		cond.MarkSuccess()
		// Update VirtualService when it's not nil. Do not overwrite it back to empty otherwise.
		if vs != nil {
			status.VirtualService = corev1.LocalObjectReference{Name: vs.Name}
		}
	case vsErr != nil:
		cond.MarkReconciliationError("reconciling", vsErr)
		status.VirtualService = corev1.LocalObjectReference{}
	case vs == nil:
		cond.MarkReconciliationPending()
		status.VirtualService = corev1.LocalObjectReference{}

	default:
		cond.MarkSuccess()
		status.VirtualService = corev1.LocalObjectReference{Name: vs.Name}
	}
}

// PropagateRouteSpecFields stores the RouteSpecFields in the
// RouteStatus.
func (status *RouteStatus) PropagateRouteSpecFields(b RouteSpecFields) {
	status.RouteSpecFields = b
}

// PropagateBindings updates the list of bindings.
func (status *RouteStatus) PropagateBindings(bindings []RouteDestination) {
	if len(bindings) > 0 {
		status.Bindings = bindings
	} else {
		status.Bindings = nil
	}

	appNames := sets.NewString()
	for _, binding := range bindings {
		appNames.Insert(binding.ServiceName)
	}
	if appNames.Len() > 0 {
		status.AppBindingDisplayNames = appNames.List() // returns a sorted list
	} else {
		status.AppBindingDisplayNames = nil
	}
}

// PropagateRouteServiceBinding updates the route service name on the Route.
func (status *RouteStatus) PropagateRouteServiceBinding(routeServices []RouteServiceDestination) {
	if len(routeServices) > 1 {
		var routeServiceNames []string
		for _, routeService := range routeServices {
			routeServiceNames = append(routeServiceNames, routeService.Name)
		}
		status.manage().MarkFalse(RouteConditionRouteServiceReady, "MultipleRouteServices", "More than one route service is bound: [%s]",
			strings.Join(routeServiceNames, ", "))

		// Set the RouteService name in the status to the most recently bound service.
		// This essentially acts as a placeholder since the condition is already set to False.
		status.RouteService = corev1.LocalObjectReference{
			Name: routeServices[len(routeServices)-1].Name,
		}
		return
	}
	if len(routeServices) > 0 {
		status.RouteService = corev1.LocalObjectReference{
			Name: routeServices[0].Name,
		}
	} else {
		status.RouteService = corev1.LocalObjectReference{}
	}
	status.RouteServiceCondition().MarkSuccess()
}

// PropagateSpaceDomain sets status fields from the given SpaceDomain. A nil
// SpaceDomain is interpreted as not being defined (and thus permitted) on
// the Space.
func (status *RouteStatus) PropagateSpaceDomain(spaceDomain *SpaceDomain) {
	switch {
	case spaceDomain == nil:
		status.SpaceDomainCondition().MarkReconciliationError("InvalidDomain", errors.New("The domain specified on the Route isn't permitted by the Space"))
	default:
		status.SpaceDomainCondition().MarkSuccess()
	}
}
