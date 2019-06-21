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
import duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
import "k8s.io/kubernetes/pkg/apis/core"

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

	// Source contains the source configuration of the App.
	// +optional
	Source AppSpecSource `json:"source,omitempty"`

	// Template defines the App's desired runtime configuration.
	// +optional
	Template AppSpecTemplate `json:"template,omitempty"`

	// Routes defines network routes for the App's ingress.
	// +optional
	Routes AppSpecRoutes `json:"routes,omitempty"`

	// Services defines what services the App requires.
	// +optional
	Services AppSpecServices `json:"services,omitempty"`

	Instances AppSpecInstances `json:"instances,omitempty"`
}

// AppSpecSource defines the source code for an App.
// The fields ContainerImage and BuildpackBuild are mutually exclusive.
type AppSpecSource struct {

	// UId is a unique identifier for an AppSpecSource.
	// Updating sub-values will trigger a new value.
	// +optional
	UId string `json:"uid,omitempty"`

	// ContainerImage defines the container image for source.
	// +optional
	ContainerImage AppSpecSourceContainerImage `json:"containerimage,omitempty"`

	// BuildpackBuild defines buildpack information for source.
	// +optional
	BuildpackBuild AppSpecSourceBuildpackBuild `json:"buildpackbuild,omitempty"`
}

// AppSpecSourceContainerImage defines a container image for an App.
type AppSpecSourceContainerImage struct {

	// Image is the container image URI for the App.
	Image string `json:"image"`
}

// AppSpecSourceBuildpackBuild defines building an App using Buildpacks.
type AppSpecSourceBuildpackBuild struct {

	// Source is the Container Image which contains the App's source code.
	// +optional
	Source string `json:"source,omitempty"`

	// Stack is the base layer to use for the App.
	// +optional
	Stack string `json:"stack,omitempty"`

	// Buildpack is the Buildpack to use for the App.
	// +optional
	Buildpack string `json:"buildpack,omitempty"`
}

// AppSpecTemplate defines an App's desired runtime configuration.
type AppSpecTemplate struct {

	// UId is a unique identifier for an AppSpecTemplate.
	// Updating sub-values will trigger a new value.
	// +optional
	UId string `json:"uid,omitempty"`

	// Spec is a PodSpec with additional restrictions.
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

type AppSpecInstances struct {
	stopped *bool `json:"stopped,omitempty"`
	exactly *int  `json:"exactly,omitempty"`
	min     *int  `json:"min,omitempty"`
	max     *int  `json:"max,omitempty"`
}

// AppStatus is the current configuration and running state for an App.
type AppStatus struct {

	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList is a life of App resources.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}
