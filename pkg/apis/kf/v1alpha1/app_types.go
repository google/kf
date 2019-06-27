// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
import core "k8s.io/api/core/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// App is a 12-factor application deployed to Knative. It encompasses source
// code, configuration, and the current state of the application.
type App struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec AppSpec `json:"spec,omitempty"`

	// +optional
	Status AppStatus `json:"status,omitempty"`
}

// AppSpec is the desired configuration for an App.
type AppSpec struct {

	// SourceSpec contains the source configuration of the App.
	// +optional
	SourceSpec `json:"source,inline"`

	// Template defines the App's runtime configuration.
	// +optional
	Template AppSpecTemplate `json:"template"`

	// Routes defines network routes for the App's ingress.
	// +optional
	Routes AppSpecRoutes `json:"routes,omitempty"`

	// Services defines what services the App requires.
	// +optional
	Services AppSpecServices `json:"services,omitempty"`

	// Instances defines the scaling rules for the App.
	Instances AppSpecInstances `json:"instances,omitempty"`
}

// AppSpecTemplate defines an app's runtime configuration.
type AppSpecTemplate struct {

	// UpdateRequests is a unique identifier for an AppSpecTemplate.
	// Updating sub-values will trigger a new value.
	UpdateRequests int `json:"updateRequests"`

	// Template is a PodSpec with additional restrictions.
	// The image name is ignored.
	// The Spec contains configuration for the App's Pod.
	// (Env, Vars, Quotas, etc)
	// +optional
	Spec core.PodSpec `json:"spec,omitempty"`
}

// AppSpecRoutes defines network routes for an App's ingress.
type AppSpecRoutes struct {
}

// AppSpecServices defines what services an App requires.
type AppSpecServices struct {
}

// AppSpecInstances defines the scaling rules for an App.
type AppSpecInstances struct {

	// Stopped determines if the App should be running or not.
	Stopped bool `json:"stopped,omitempty"`

	// Exactly defines a static number of desired instances.
	// If Exactly is set, it supersedes the Min and Max values.
	Exactly *int `json:"exactly,omitempty"`

	// Min defines a minimum auto-scaling limit.
	Min *int `json:"min,omitempty"`

	// Max defines a maximum auto-scaling limit.
	Max *int `json:"max,omitempty"`
}

// AppStatus is the current configuration and running state for an App.
type AppStatus struct {

	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList is a list of App resources.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}
