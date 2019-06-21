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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Source represents the source code and build configuration for an App.
type Source struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec SourceSpec `json:"spec,omitempty"`

	// +optional
	Status SourceStatus `json:"status,omitempty"`
}

// SourceSpec defines the source code for an App.
// The fields ContainerImage and BuildpackBuild are mutually exclusive.
type SourceSpec struct {

	// UpdateRequests is a unique identifier for an AppSpecSource.
	// Updating sub-values will trigger a new value.
	UpdateRequests int `json:"updaterequests"`

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

// SourceStatus is the current configuration and running state for an App's Source.
type SourceStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourceList is a life of Source resources.
type SourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Source `json:"items"`
}
