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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import corev1 "k8s.io/api/core/v1"
import duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Space is a high level structure that encompasses a namespace, permissions on
// it and configuration applied to it.
type Space struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec SpaceSpec `json:"spec,omitempty"`

	// +optional
	Status SpaceStatus `json:"status,omitempty"`
}

// SpaceSpec contains the specification for a space.
type SpaceSpec struct {
	// Security contains config for RBAC roles that will be created for the
	// space.
	// +optional
	Security SpaceSpecSecurity `json:"security,omitempty"`

	// BuildpackBuild contains config for the build pipelines.
	// Currently, this is the only way to build source -> container workflows, but
	// in the future additional types may be added. For example DockerBuild or
	// WebhookBuild to execute a build via webhook.
	// +optional
	BuildpackBuild SpaceSpecBuildpackBuild `json:"buildpackBuild,omitempty"`

	// Execution contains settings for the execution environment.
	// +optional
	Execution SpaceSpecExecution `json:"execution,omitempty"`

	// SpaceSpecResourceLimits contains definitions for resource usage limits.
	// +optional
	ResourceLimits SpaceSpecResourceLimits `json:"resourceLimits,omitempty"`
}

// SpaceSpecSecurity holds fields for creating RBAC in the space.
type SpaceSpecSecurity struct {
	// NOTE: The false value for each field should be the default and safe.

	// EnableDeveloperLogsAccess allows developers to access pod logging endpoints.
	// +optional
	EnableDeveloperLogsAccess bool `json:"enableDeveloperLogsAccess,omitempty"`
}

// SpaceSpecBuildpackBuild holds fields for managing building via buildpacks.
type SpaceSpecBuildpackBuild struct {
	// NOTE: The false value for each field should be the default and safe.

	// BuilderImage is a buildpacks.io builder image.
	BuilderImage string `json:"builderImage,omitempty"`

	// ContainerRegistry holds the container registry that buildpack builds are
	// stored in.
	ContainerRegistry string `json:"containerRegistry,omitempty"`

	// Env sets default environment variables on the builder.
	//
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// SpaceSpecExecution contains settings for the execution environment.
type SpaceSpecExecution struct {
	// Env sets default environment variables on kf applications for the whole
	// space.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Domains sets valid domains that can be used for routes in the space.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Domains []SpaceDomain `json:"domains,omitempty" patchStrategy:"merge" patchMergeKey:"domain"`
}

// SpaceSpecResourceLimits contains definitions for resource usage limits.
type SpaceSpecResourceLimits struct {
	// SpaceQuota holds the k8s ResourceQuota created for the whole space.
	// For now, only one ResourceQuota per space is supported.
	// Consider allowing multiple ResourceQuotas when more quota scopes are enabled in k8s
	// +optional
	SpaceQuota corev1.ResourceList `json:"spaceQuota,omitempty"`

	// ResourceDefaults holds the k8s LimitRange created for the whole space,
	// which sets default request/limit for resources per pod or container.
	// +optional
	ResourceDefaults []corev1.LimitRangeItem `json:"resourceDefaults,omitempty"`
}

// SpaceDomain stores information about a domain available in a space.
type SpaceDomain struct {
	// Domain is the valid domain that can be used in conjunction with a
	// hostname and path for a route.
	Domain string `json:"domain"`

	// Default implies that this SpaceDomain is used when a domain is not
	// specified. There can only be a single default set to true per space.
	// NOTE: This may change in the future.
	Default bool `json:"default,omitempty"`
}

// SpaceStatus represents information about the status of a Space.
type SpaceStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	Quota corev1.ResourceQuotaStatus `json:"quota,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpaceList is a list of KfSpace resources
type SpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Space `json:"items"`
}
