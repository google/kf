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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Route is a high level structure that encompasses an Istio VirtualService
// and configuration applied to it.
type Route struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RouteSpec `json:"spec,omitempty"`

	// +optional
	Status RouteStatus `json:"status,omitempty"`
}

// RouteSpec contains the specification for a route.
type RouteSpec struct {
	// KnativeServiceNames contains the Kf Apps that are bound to the route.
	// +optional
	KnativeServiceNames []string `json:"knativeServiceNames"`

	// Hostname is the hostname or subdomain of the route (e.g, in
	// hostname.example.com it would be hostname).
	// +optional
	Hostname string

	// Domain is the domain of the route (e.g, in hostname.example.com it
	// would be example.com).
	// +optional
	Domain string

	// Path is the URL path of the route.
	// +optional
	Path string
}

// RouteStatus represents information about the status of a Route.
type RouteStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteList is a list of Route resources
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Route `json:"items"`
}
