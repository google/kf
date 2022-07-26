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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
)

const (
	VolumeBrokerKind = "VolumeBroker"
)

// CommonServiceBroker is an interface common to cluster and namespaced brokers.
type CommonServiceBroker interface {
	// GetName returns the name of the broker.
	GetName() string
	// GetNamespace returns the name of the namespace or blank if cluster-scoped.
	GetNamespace() string
	// GetKind returns the kind of the broker.
	GetKind() string
	// GetServiceOfferings returns the service offerings for the broker.
	GetServiceOfferings() []ServiceOffering
	// GetCredentialsSecretRef gets the Secret reference for the connection
	// credentials.
	GetCredentialsSecretRef() NamespacedObjectReference
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBroker represents an Open Service Broker (OSB) compatible service
// broker.
type ServiceBroker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBrokerSpec         `json:"spec,omitempty"`
	Status CommonServiceBrokerStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*ServiceBroker)(nil)
var _ apis.Defaultable = (*ServiceBroker)(nil)
var _ CommonServiceBroker = (*ServiceBroker)(nil)
var _ kmeta.OwnerRefable = (*ServiceBroker)(nil)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (sb *ServiceBroker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceBroker")
}

// GetServiceOfferings implements CommonServiceBroker.
func (sb *ServiceBroker) GetKind() string {
	return sb.Kind
}

// GetServiceOfferings implements CommonServiceBroker.
func (sb *ServiceBroker) GetServiceOfferings() []ServiceOffering {
	return sb.Status.Services
}

// GetCredentialsSecretRef implements CommonServiceBroker.
func (sb *ServiceBroker) GetCredentialsSecretRef() NamespacedObjectReference {
	return NamespacedObjectReference{
		Namespace: sb.Namespace,
		Name:      sb.Spec.Credentials.Name,
	}
}

// ServiceBrokerSpec contains the user supplied specification for the broker.
type ServiceBrokerSpec struct {
	CommonServiceBrokerSpec `json:",inline"`

	// Credentials contains a reference to a secret containing credentials
	// for the service.
	// +optional
	Credentials corev1.LocalObjectReference `json:"credentials"`
}

// VolumeBrokerSpec contains the user supplied specification for the broker.
type VolumeBrokerSpec struct {
	// VolumeOfferings contains ServiceOfferings supported by this broker.
	VolumeOfferings []ServiceOffering `json:"offering,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBroker represents an Open Service Broker (OSB) compatible
// service broker available at the cluster level.
type ClusterServiceBroker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterServiceBrokerSpec  `json:"spec,omitempty"`
	Status CommonServiceBrokerStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*ClusterServiceBroker)(nil)
var _ apis.Defaultable = (*ClusterServiceBroker)(nil)
var _ CommonServiceBroker = (*ClusterServiceBroker)(nil)
var _ kmeta.OwnerRefable = (*ClusterServiceBroker)(nil)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (sb *ClusterServiceBroker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ClusterServiceBroker")
}

// GetKind implements CommonServiceBroker.
func (sb *ClusterServiceBroker) GetKind() string {
	if sb.Spec.CommonServiceBrokerSpec.VolumeBrokerSpec != nil {
		return VolumeBrokerKind
	}
	return sb.Kind
}

// GetServiceOfferings implements CommonServiceBroker.
func (sb *ClusterServiceBroker) GetServiceOfferings() []ServiceOffering {
	return sb.Status.Services
}

// GetCredentialsSecretRef implements CommonServiceBroker.
func (sb *ClusterServiceBroker) GetCredentialsSecretRef() NamespacedObjectReference {
	return NamespacedObjectReference{
		Namespace: sb.Spec.Credentials.Namespace,
		Name:      sb.Spec.Credentials.Name,
	}
}

// ClusterServiceBrokerSpec contains the user supplied specification for the broker.
type ClusterServiceBrokerSpec struct {
	CommonServiceBrokerSpec `json:",inline"`

	// Credentials contains a reference to a secret containing credentials
	// for the service.
	// +optional
	Credentials NamespacedObjectReference `json:"credentials"`
}

// NamespacedObjectReference is like corev1.LocalObjectReference but includes
// a Namespace specifier.
type NamespacedObjectReference struct {
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	Namespace string `json:"namespace"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`
}

// CommonServiceBrokerSpec holds common fields between the ServiceBrokerSpec
// and ClusterServiceBrokerSpec.
type CommonServiceBrokerSpec struct {
	// UpdateRequests is a unique identifier, updating will trigger a
	// refresh.
	// +optional
	UpdateRequests int `json:"updateRequests"`

	// Future additions may need to be made to enable operators to enable/disable
	// services in particular namespaces; these fields should be inlined here
	// a shared structure between ServiceBroker and ClusterServiceBroker.

	// VolumeBrokerSpec indicates this service broker is a VolumeBroker.
	VolumeBrokerSpec *VolumeBrokerSpec `json:"volume,omitempty"`
}

// CommonServiceBrokerStatus contains the status of the broker.
type CommonServiceBrokerStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// UpdateRequests is the last processed UpdateRequests value.
	UpdateRequests int `json:"updateRequests"`

	// Services contains the list of services offered by the broker.
	Services []ServiceOffering `json:"services,omitempty"`
}

// ServiceOffering has just enough info to display the offering in
// the marketplace command and provision it.
type ServiceOffering struct {
	// DisplayName is the human readable name of the offering, the
	// field is unstable across releases.
	DisplayName string `json:"displayName"`
	// UID is a unique ID of the offering within the broker, this value is stable
	// across broker releases and is used to track when names change.
	// It's recommended, but not required that this value be a UUID.
	UID string `json:"uid"`
	// Description is a human readable description of the offering.
	Description string `json:"description"`
	// Tags contains opaque labels to help filter marketplace, examples include:
	// gcp, sql, myssql.
	// +nullable
	Tags []string `json:"tags,omitempty"`
	// Plans contains a list of tiers that can be provisioned. For example,
	// databases might come in different sizes.
	Plans []ServicePlan `json:"plans,omitempty"`
}

// ServicePlan has just enough info to display the offering in
// the marketplace command and provision it.
type ServicePlan struct {
	// DisplayName is the human readable name of the plan. This value is unstable
	// across releases.
	DisplayName string `json:"displayName"`
	// Free indicates that the plan has no cost to the end-user.
	// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#service-plan-object
	Free bool `json:"free"`
	// UID is the unique ID of the plan (within the service). The value is
	// stable across broker releases.
	// It's recommended, but not required that this value be a UUID.
	UID string `json:"uid"`
	// Description is a human readable description of the plan.
	Description string `json:"description"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBrokerList is a list of ServiceBroker resources
type ServiceBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceBroker `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBrokerList is a list of ClusterServiceBroker resources
type ClusterServiceBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterServiceBroker `json:"items"`
}
