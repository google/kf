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

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

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

	// UpdateRequests is a unique identifier for an SourceSpec.
	// Updating sub-values will trigger a new value.
	// +optional
	UpdateRequests int `json:"updateRequests,omitempty"`

	// ServiceAccount holds the service account the build will execute as.
	// This can be used to attach credentials to the build.
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// ContainerImage defines the container image for source.
	// +optional
	ContainerImage SourceSpecContainerImage `json:"containerImage,omitempty"`

	// BuildpackBuild defines buildpack information for source.
	// +optional
	BuildpackBuild SourceSpecBuildpackBuild `json:"buildpackBuild,omitempty"`

	// Dockerfile defines Dockerfile information for source.
	// +optional
	Dockerfile SourceSpecDockerfile `json:"dockerfile,omitempty"`
}

// NeedsUpdateRequestsIncrement returns true if UpdateRequests needs to be
// incremented to force a rebuild. This function should be used as a part of
// defaulting and validating webhooks when SourceSpec is embedded.
//
// This can happen if a field in the source changes without also updating the
// UpdateRequests.
func (spec *SourceSpec) NeedsUpdateRequestsIncrement(old SourceSpec) bool {
	updateRequestsChanged := old.UpdateRequests != spec.UpdateRequests

	if !updateRequestsChanged {
		specsChanged := !reflect.DeepEqual(&old, spec)
		return specsChanged
	}

	return false
}

// SourceSpecContainerImage defines a container image for an App.
type SourceSpecContainerImage struct {

	// Image is the container image URI for the App.
	Image string `json:"image"`
}

// SourceSpecBuildpackBuild defines building an App using Buildpacks.
type SourceSpecBuildpackBuild struct {

	// Source is the Container Image which contains the App's source code.
	Source string `json:"source"`

	// Stack is the base layer to use for the App.
	// +optional
	Stack string `json:"stack,omitempty"`

	// Buildpack is the Buildpack to use for the App.
	// +optional
	Buildpack string `json:"buildpack,omitempty"`

	// BuildpackBuilder is the container image which builds the App.
	BuildpackBuilder string `json:"buildpackBuilder"`

	// Image is the location to store the built image.
	Image string `json:"image"`

	// Env represents the environment variables to apply when building the App.
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// SourceSpecDockerfile defines building an App using a Dockerfile.
type SourceSpecDockerfile struct {

	// Source is the Container Image which contains the App's source code and
	// Dockerfile
	Source string `json:"source"`

	// Path is the path to the dockerfile.
	Path string `json:"path"`

	// Image is the location to store the built image.
	Image string `json:"image"`
}

// SourceStatus is the current configuration and running state for an App's Source.
type SourceStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	SourceStatusFields `json:",inline"`
}

// SourceStatusFields holds the fields of Source's status that
// are shared. This is defined separately and inlined so that
// other types can readily consume these fields via duck typing.
type SourceStatusFields struct {
	// Image is the latest successfully built image.
	// +optional
	Image string `json:"image,omitempty"`

	// BuildName is the name of the build that produced the image.
	// +optional
	BuildName string `json:"buildName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourceList is a list of Source resources.
type SourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Source `json:"items"`
}

// IsContainerBuild returns true if the build is for a container
func (spec *SourceSpec) IsContainerBuild() bool {
	return spec.ContainerImage.Image != ""
}

// IsBuildpackBuild returns true if the build is for a buildpack
func (spec *SourceSpec) IsBuildpackBuild() bool {
	return spec.BuildpackBuild.Source != ""
}

// IsDockerfileBuild returns true if the build is for a dockerfile
func (spec *SourceSpec) IsDockerfileBuild() bool {
	return spec.Dockerfile.Source != ""
}
