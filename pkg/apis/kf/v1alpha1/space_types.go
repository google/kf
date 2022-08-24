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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// NetworkPolicyLabel holds a label key for determining which default network
	// policy will be applied to a Pod.
	NetworkPolicyLabel = "kf.dev/networkpolicy"

	// NetworkPolicyApp holds the NetworkPolicyLabel value for app policies.
	NetworkPolicyApp = "app"

	// NetworkPolicyBuild holds the NetworkPolicyLabel value for build policies.
	NetworkPolicyBuild = "build"

	// PermitAllNetworkPolicy is the key used to indcate all traffic is allowed.
	PermitAllNetworkPolicy = "PermitAll"
	// DenyAllNetworkPolicy is the key used to indcate all traffic is denied.
	DenyAllNetworkPolicy = "DenyAll"
)

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

var _ apis.Validatable = (*Space)(nil)
var _ apis.Defaultable = (*Space)(nil)

// DefaultDomainOrBlank gets the default domain for the space if set, otherwise
// blank.
func (s *Space) DefaultDomainOrBlank() string {
	for _, domain := range s.Status.NetworkConfig.Domains {
		return domain.Domain
	}

	return ""
}

// SpaceSpec contains the specification for a space.
type SpaceSpec struct {
	// BuildConfig contains config for the build pipelines.
	// +optional
	BuildConfig SpaceSpecBuildConfig `json:"buildConfig,omitempty"`

	// RuntimeConfig contains settings for the app runtime environment.
	// +optional
	RuntimeConfig SpaceSpecRuntimeConfig `json:"runtimeConfig,omitempty"`

	// NetworkConfig contains settings for the space's networking environment.
	// +optional
	NetworkConfig SpaceSpecNetworkConfig `json:"networkConfig,omitempty"`
}

// SpaceSpecBuildConfig holds fields for managing building.
type SpaceSpecBuildConfig struct {
	// ContainerRegistry holds the container registry that buildpack builds are
	// stored in.
	ContainerRegistry string `json:"containerRegistry,omitempty"`

	// Env sets default environment variables on the builder.
	//
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// DefaultToV3Stack tells kf whether it applications should default
	// to using V3 stacks. If not supplied, the default value is taken from the
	// config-defaults configmap.
	//
	// +nullable
	// +optional
	DefaultToV3Stack *bool `json:"defaultToV3Stack"`

	// ServiceAccount is the service account that will be propagated to
	// all builds.
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// SpaceSpecRuntimeConfig contains config for the actual applciation runtime
// configuration for the space.
type SpaceSpecRuntimeConfig struct {
	// Env sets default environment variables on kf applications for the whole
	// space.
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// NodeSelector sets the NodeSelector in the podSpec to invoke Node Assignment
	// https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// SpaceSpecNetworkConfig contains settings for the space's networking.
type SpaceSpecNetworkConfig struct {
	// Domains sets valid domains that can be used for routes in the space.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Domains []SpaceDomain `json:"domains,omitempty" patchStrategy:"merge" patchMergeKey:"domain"`

	// AppNetworkPolicy holds the default network policy for apps.
	AppNetworkPolicy SpaceSpecNetworkConfigPolicy `json:"appNetworkPolicy,omitempty"`

	// BuildNetworkPolicy holds the default network policy for builds.
	BuildNetworkPolicy SpaceSpecNetworkConfigPolicy `json:"buildNetworkPolicy,omitempty"`
}

// SpaceSpecNetworkConfigPolicy holds the policy for a particular type.
type SpaceSpecNetworkConfigPolicy struct {
	// Ingress holds the default network policy for inbound traffic.
	Ingress string `json:"ingress,omitempty"`
	// Egress holds the default network policy for outbound traffic.
	Egress string `json:"egress,omitempty"`
}

// SpaceDomain stores information about a domain available in a space.
type SpaceDomain struct {
	// Domain is the valid domain that can be used in conjunction with a
	// hostname and path for a route.
	Domain string `json:"domain"`

	// GatewayName is the name of the Istio Gateway supported by the domain.
	// Values can include a Namespace as a prefix.
	// Only the kf Namespace is allowed e.g. kf/some-gateway.
	// See https://istio.io/docs/reference/config/networking/gateway/
	GatewayName string `json:"gatewayName,omitempty"`
}

// StableDeduplicateSpaceDomainList removes SpaceDomain with duplicate domain fields
// preserving the order of the input.
func StableDeduplicateSpaceDomainList(domains []SpaceDomain) (out []SpaceDomain) {
	known := sets.NewString()
	for _, domain := range domains {
		if known.Has(domain.Domain) {
			continue
		}

		known.Insert(domain.Domain)
		out = append(out, domain)
	}

	return
}

// SpaceStatus represents information about the status of a Space.
type SpaceStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// RuntimeConfig contains the info necessary to configure the application
	// runtime.
	RuntimeConfig SpaceStatusRuntimeConfig `json:"runtimeConfig,omitempty"`

	// NetworkConfig contains the info necessary to configure application
	// networking.
	NetworkConfig SpaceStatusNetworkConfig `json:"networkConfig,omitempty"`

	// BuildConfig contains the info necessary to configure builds.
	BuildConfig SpaceStatusBuildConfig `json:"buildConfig,omitempty"`

	// IngressGateways contains the list of ingress gateways that could
	// direct traffic into this Kf space.
	IngressGateways []corev1.LoadBalancerIngress `json:"ingressGateways"`
}

// FindIngressIP gets the lexicographicaly first IP address from a the
// listed ingress gateways if one exists.
func (s *SpaceStatus) FindIngressIP() *string {
	ips := sets.NewString()
	for _, gateway := range s.IngressGateways {
		if ip := gateway.IP; ip != "" {
			ips.Insert(ip)
		}
	}

	// List returns a sorted list.
	orderedIPs := ips.List()

	if len(orderedIPs) > 0 {
		return &orderedIPs[0]
	}

	return nil
}

// SpaceStatusRuntimeConfig reflects the actual applciation runtime
// configuration for the space.
type SpaceStatusRuntimeConfig struct {
	// Env sets default environment variables on kf applications for the whole
	// space.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// AppCPUPerGBOfRAM sets the default amount of CPU to assign an app per GB of RAM.
	AppCPUPerGBOfRAM *resource.Quantity `json:"appCPUPerGBOfRAM,omitempty"`

	// AppCPUMin sets the minimum amount of CPU to assign an app regardless of the
	// amount of RAM it's assigned.
	AppCPUMin *resource.Quantity `json:"appCPUMin,omitempty"`

	// ProgressDeadlineSeconds contains a configurable timeout between state 
	// transition and reaching a stable state before provisioning or deprovisioning
	// times out.
	ProgressDeadlineSeconds int32 `json:"progressDeadlineSeconds,omitempty"`
}

// SpaceStatusNetworkConfig reflects the actual Networking configuration for the
// space.
type SpaceStatusNetworkConfig struct {
	// Domains sets valid domains that can be used for routes in the space.
	// +optional
	Domains []SpaceDomain `json:"domains,omitempty"`
}

// SpaceStatusBuildConfig reflects the actual build configuration for the
// space.
type SpaceStatusBuildConfig struct {
	// BuildpacksV2 contains a list of V2 (Cloud Foundry) compatible buildpacks
	// that will be available by builders in the space.
	BuildpacksV2 config.BuildpackV2List `json:"buildpacksV2,omitempty"`

	// StacksV2 contains a list of V2 (Cloud Foundry) compatible stacks
	// that will be available by builders in the space.
	StacksV2 config.StackV2List `json:"stacksV2,omitempty"`

	// StacksV3 contains a list of V3 (Cloud Native Buildpacks) compatible stacks
	// that will be available by builders in the space.
	StacksV3 config.StackV3List `json:"stacksV3,omitempty"`

	// Env contains additional build environment variables for the whole space.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// ServiceAccount is the service account that will be propagated to
	// all builds.
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// ContainerRegistry holds the container registry that buildpack builds are
	// stored in.
	ContainerRegistry string `json:"containerRegistry,omitempty"`

	// DefaultToV3Stack tells kf whether it applications should default to using
	// V3 stacks.
	DefaultToV3Stack bool `json:"defaultToV3Stack"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpaceList is a list of KfSpace resources
type SpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Space `json:"items"`
}
