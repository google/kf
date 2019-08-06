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
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:noStatus

// Route is a high level structure that encompasses an Istio VirtualService
// and configuration applied to it.
type Route struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RouteSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteList is a list of Route resources
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Route `json:"items"`
}

// RouteSpec contains the specification for a Route.
type RouteSpec struct {
	// AppName contains the Kf App that is bound to the route.
	AppName string `json:"appName,omitempty"`

	// RouteSpecFields contains the fields of a route.
	RouteSpecFields `json:",inline"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:noStatus

// RouteClaim is similar to Route, however it is not associated with an App.
// It is created (by the Route Controller) along with its associated Routes.
type RouteClaim struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RouteClaimSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteClaimList is a list of RouteClaim resources
type RouteClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RouteClaim `json:"items"`
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

// String returns a RouteSpecFields converted into an address.
func (route RouteSpecFields) String() string {
	var hostnamePrefix string
	if route.Hostname != "" {
		hostnamePrefix = route.Hostname + "."
	}
	return hostnamePrefix + route.Domain + path.Join("/", route.Path)
}

// RouteClaimSpec contains the specification for a RouteClaim.
type RouteClaimSpec struct {
	// RouteSpecFields contains the fields of a route.
	RouteSpecFields `json:",inline"`
}
