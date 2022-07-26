// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	kfdefaults "kf-operator/pkg/apis/kfsystem/kf"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// KfInstallSucceeded is condition type for successful Kf installation.
	KfInstallSucceeded apis.ConditionType = "KfInstallSucceeded"
)

var (
	_ duckv1.KRShaped = (*KfSystem)(nil)
)

// KfSpec defines the desired state of Kf.
type KfSpec struct {
	Enabled *bool `json:"enabled,omitempty"`

	Version string `json:"version,omitempty"`

	// Config defines the configurations of Kf installation
	Config KfConfig `json:"config,omitempty"`
}

// KfConfig defines custom configurations for kf
type KfConfig struct {
	// SecretSpec defines the secrets for the cluster.
	Secrets SecretSpec `json:"secrets,omitempty"`

	// TODO: Switch to use imported Kf types
	// DefaultsConfig contains the configuration for defaults.
	kfdefaults.DefaultsConfig

	// GatewaySpec defines the configuration for the Kubernetes Gateway
	Gateway *GatewaySpec `json:"gateway,omitempty"`
}

// SecretSpec defines the secrets for the cluster.
type SecretSpec struct {
	// ControllerCACerts points to a secret that stores the CA certs that are
	// mounted to the controller pods. Each key in the secret is considered
	// a file in /etc/ssl/certs. The secret MUST be immutable.
	// +optional
	ControllerCACerts corev1.LocalObjectReference `json:"controllerCACerts,omitempty"`

	// WorkloadIdentity configures the cluster to use Workload Identity. One
	// of WorkloadIdentity or Build must be set (but not both).
	// +optional
	WorkloadIdentity *SecretWorkloadIdentity `json:"workloadidentity,omitempty"`

	// Build configures the cluster to use the given Secret to interact with a
	// non-GCP container registry. One of WorkloadIdentity or Build must be
	// set (but not both).
	// +optional
	Build *SecretBuild `json:"build,omitempty"`
}

// SecretWorkloadIdentity defines the configuration for Workload Identity.
type SecretWorkloadIdentity struct {
	// GoogleServiceAccount is the GSA that will be used for the token exchange.
	GoogleServiceAccount string `json:"googleserviceaccount,omitempty"`

	// GoogleProjectID is the ProjectID that the GSA exists in.
	GoogleProjectID string `json:"googleprojectid,omitempty"`
}

// SecretBuild defines the configuration for the Kubernetes Secret that holds the
// auth for interacting with a non-GCP container registry.
type SecretBuild struct {
	// ImagePushSecrets is the name of the Kubernetes Secret that user created in the Kf namespace
	ImagePushSecretName string `json:"imagePushSecrets,omitempty"`
}

// GatewaySpec defines the configuration for the Kubernetes Gateway
// Cluster admin would create the Gateway in a Namespace and Kf Operator will point
// to the Gateway through the GatewaySpec
type GatewaySpec struct {
	// IngressGateway provides a means to override the ingressgateway
	IngressGateway IstioGatewayOverride `json:"ingressgateway,omitempty"`

	// ClusterLocalGateway provides a means to override the clusterlocalgateway
	ClusterLocalGateway IstioGatewayOverride `json:"clusterlocalgateway,omitempty"`
}

// IstioGatewayOverride override the ingress-gateway and cluster-local-gateway
type IstioGatewayOverride struct {
	// Selector defines a map of values to replace the "selector" values in the
	// ingress-gateway and cluster-local-gateway
	Selector map[string]string `json:"selector,omitempty"`
}

// KfStatus defines the observed state of Kf.
type KfStatus struct {
	duckv1.Status `json:",inline"`

	// The version of the installed release
	// +optional
	KfVersion string `json:"kfversion,omitempty"`
}

// KfSystemSpec defines the desired state of KfSystem.
type KfSystemSpec struct {
	Kf KfSpec `json:"kf,omitempty"`
}

// KfSystemStatus defines the observed state of KfSystem.
type KfSystemStatus struct {
	duckv1.Status `json:",inline"`

	// The version of the installed release
	// +optional
	KfVersion string `json:"kfversion,omitempty"`

	// The targeted version of kf release
	// +optional
	TargetKfVersion string `json:"targetkfversion,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KfSystem is the Schema for the KfSystems API.
type KfSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KfSystemSpec   `json:"spec,omitempty"`
	Status KfSystemStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KfSystemList contains a list of KfSystems.
type KfSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KfSystem `json:"items"`
}
