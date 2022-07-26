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
	"context"
	"net/url"
	"path"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/ptr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Route is a mapping between a Hostname/Domain/Path combination and Apps
// that want to receive traffic from it.
type Route struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RouteSpec `json:"spec,omitempty"`

	// +optional
	Status RouteStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*Route)(nil)
var _ apis.Defaultable = (*Route)(nil)

// IsOrphaned returns true if the route is orphaned or false if it is not.
// Routes that are out of sync are assumed not to be orphaned.
func (r *Route) IsOrphaned() bool {
	// If the object is out of sync or hasn't been reconciled yet
	// don't assume it's orphaned.
	if r.Generation == 0 || r.Generation != r.Status.ObservedGeneration {
		return false
	}

	return len(r.Status.Bindings) == 0
}

func (r *Route) hasDestination(needle RouteDestination) bool {
	for _, dest := range r.Status.Bindings {
		if dest == needle {
			return true
		}
	}

	return false
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteList is a list of Route resources
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Route `json:"items"`
}

// RouteSpecFields contains the fields of a route.
type RouteSpecFields struct {
	// Hostname is the hostname or subdomain of the route (e.g, in
	// hostname.example.com it would be hostname).
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// Domain is the domain of the route (e.g, in hostname.example.com it
	// would be example.com).
	// +optional
	Domain string `json:"domain,omitempty"`

	// Path is the URL path of the route.
	// +optional
	Path string `json:"path,omitempty"`
}

// RouteWeightBinding contains the fields of a route.
type RouteWeightBinding struct {
	// RouteSpecFields contains the fields of a route.
	RouteSpecFields `json:",inline"`

	// Weight is the weight of the app in the route.
	// Every app has a default weight of 1, meaning if there are multiple apps
	// mapped to a route, traffic will be uniformly distributed among them.
	// If an app is stopped, its weight is 0.
	// +optional
	Weight *int32 `json:"weight"`

	// DestinationPort contains the port number of the service the Route
	// connects to. It can only be nil on Apps or for routes that were created
	// before it was a required field, in which case it will be defaulted to 80
	// at runtime.
	DestinationPort *int32 `json:"destinationPort,omitempty"`
}

// Merge adds the weights of two RouteWeightBindings that are equal.
func (rwb *RouteWeightBinding) Merge(other RouteWeightBinding) {
	var weight int32 = 0
	if rwb.Weight != nil {
		weight += *rwb.Weight
	} else {
		weight += defaultRouteWeight
	}

	if other.Weight != nil {
		weight += *other.Weight
	} else {
		weight += defaultRouteWeight
	}

	rwb.Weight = &weight
}

// MergeBindings merges bindings that point to the same destination. They retain
// their original ordering.
//
// The algorithm is O(n^2), with few route bindings per app this should be
// workable, but if that assumption ever changes the algorithm will have to be
// made more robust.
func MergeBindings(bindings []RouteWeightBinding) (merged []RouteWeightBinding) {
	for _, binding := range bindings {
		found := false
		for i := range merged {
			if merged[i].EqualsBinding(context.Background(), binding) {
				found = true
				merged[i].Merge(binding)
				break
			}
		}

		if !found {
			merged = append(merged, *binding.DeepCopy())
		}
	}

	return
}

// MergeQualifiedBindings merges bindings that point to the same destination.
// They retain their original ordering.
//
// The algorithm is O(n^2), with few route bindings per app this should be
// workable, but if that assumption ever changes the algorithm will have to be
// made more robust.
func MergeQualifiedBindings(bindings []QualifiedRouteBinding) (merged []QualifiedRouteBinding) {
	for _, binding := range bindings {
		found := false
		for i := range merged {
			if merged[i].MergableWith(binding) {
				found = true
				merged[i].Merge(binding)
				break
			}
		}

		if !found {
			merged = append(merged, *binding.DeepCopy())
		}
	}

	return
}

// EqualsBinding tests equality between two bindings.
func (rwb *RouteWeightBinding) EqualsBinding(ctx context.Context, otherRwb RouteWeightBinding) bool {
	this := rwb.DeepCopy()
	this.SetDefaults(ctx)
	other := otherRwb.DeepCopy()
	other.SetDefaults(ctx)

	if this.DestinationPort == nil {
		this.DestinationPort = ptr.Int32(DefaultRouteDestinationPort)
	}
	if other.DestinationPort == nil {
		other.DestinationPort = ptr.Int32(DefaultRouteDestinationPort)
	}

	return this.RouteSpecFields.Equals(other.RouteSpecFields) &&
		reflect.DeepEqual(this.DestinationPort, other.DestinationPort)
}

// Qualify takes an unqualified RouteWeightBinding and turns it into a fully
// qualified one.
func (rwb *RouteWeightBinding) Qualify(defaultDomain, serviceName string) (out QualifiedRouteBinding) {
	out.Source = rwb.RouteSpecFields
	if out.Source.Domain == "" {
		out.Source.Domain = defaultDomain
	}

	out.Destination.ServiceName = serviceName
	if rwb.DestinationPort == nil {
		out.Destination.Port = DefaultRouteDestinationPort
	} else {
		out.Destination.Port = *rwb.DestinationPort
	}

	if rwb.Weight == nil {
		out.Destination.Weight = defaultRouteWeight
	} else {
		out.Destination.Weight = *rwb.Weight
	}

	return
}

// String returns a RouteSpecFields converted into an address.
func (route RouteSpecFields) String() string {
	if len(route.Path) == 0 || route.Path == "/" {
		return route.Host()
	}
	return route.Host() + path.Join("/", route.Path)
}

// IsWildcard returns whether or not the route is a wildcard e.g. *.example.com.
func (route RouteSpecFields) IsWildcard() bool {
	return route.Hostname == "*"
}

// Host returns the hostname concatenated with the domain.
func (route RouteSpecFields) Host() string {
	var hostnamePrefix string
	if route.Hostname != "" {
		hostnamePrefix = route.Hostname + "."
	}
	return hostnamePrefix + route.Domain
}

// Equals returns whether or not the route fields are equal to those of another route.
func (route RouteSpecFields) Equals(cmpRoute RouteSpecFields) bool {
	return route.String() == cmpRoute.String()
}

// ToURL creates a URL from the RouteSpecFields
func (route RouteSpecFields) ToURL() url.URL {
	return url.URL{
		Host: route.Host(),
		Path: route.Path,
	}
}

// RouteSpec contains the specification for a Route.
type RouteSpec struct {
	// RouteSpecFields contains the fields of a route.
	RouteSpecFields `json:",inline"`
}

// RouteStatus is the current configuration for a Route.
type RouteStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	RouteSpecFields `json:",inline"`

	// VirtualService is the VirtualService that is created with the Route.
	VirtualService corev1.LocalObjectReference `json:"virtualservice,omitempty"`

	// Bindings is the list of bindings the RouteSpecFields matches.
	Bindings []RouteDestination `json:"bindings,omitempty"`

	// AppBindingDisplayNames is the list of DisplayNames in the Bindings field
	// that belong to Apps.
	AppBindingDisplayNames []string `json:"appBindingDisplayNames,omitempty"`

	// RouteService is the Route Service instance bound to the route, if one exists.
	RouteService corev1.LocalObjectReference `json:"routeService,omitempty"`
}

// RouteServiceBinding represents a binding between a route and a route service.
type RouteServiceBinding struct {
	// Source is the route to listen on (the route that has a service bound to it).
	Source RouteSpecFields `json:"source"`

	// Destination is the traffic sink (a route service).
	Destination *RouteServiceURL `json:"destination"`
}

// RouteDestination represents enough information to route traffic from a source
// to a sink.
type RouteDestination struct {
	// Service is the name of the service to send traffic to.
	// With Apps, the service name is the App name.
	ServiceName string `json:"serviceName"` // always encode because blank is meaningful

	// Port is the port to send traffic to.
	Port int32 `json:"port"` // always encode because zero is meaningful

	// Weight is the proportion of traffic to send to this binding.
	Weight int32 `json:"weight"` // always encode because zero is meaningful
}

// QualifiedRouteBinding contains a fully qualified route binding with
// all fields filled.
type QualifiedRouteBinding struct {
	// Source is the route to listen on.
	Source RouteSpecFields `json:"source"` // always encode because blank is meaningful

	// Destination is the traffic sink.
	Destination RouteDestination `json:"destination"` // always encode because blank is meaningful
}

// Assert that QualifiedRouteBinding can be compared with ==
var _ map[QualifiedRouteBinding]interface{}

// ToUnqualified converts the QualifiedRouteBinding back into an unqualified
// one.
func (qrb *QualifiedRouteBinding) ToUnqualified() RouteWeightBinding {
	return RouteWeightBinding{
		RouteSpecFields: qrb.Source,
		DestinationPort: ptr.Int32(qrb.Destination.Port),
		Weight:          &qrb.Destination.Weight,
	}
}

// MergableWith returns true if the binding has matching fields across the board
// except weight.
func (qrb *QualifiedRouteBinding) MergableWith(other QualifiedRouteBinding) bool {
	return qrb.Source == other.Source &&
		qrb.Destination.Port == other.Destination.Port &&
		qrb.Destination.ServiceName == other.Destination.ServiceName
}

// Merge adds the weight of two QualifiedRouteBindings that have all other
// properties matching.
func (qrb *QualifiedRouteBinding) Merge(other QualifiedRouteBinding) {
	qrb.Destination.Weight += other.Destination.Weight
}
