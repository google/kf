// Copyright 2020 Google LLC
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
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// ServiceInstanceParamsSecretKey contains the secret key that holds parameters.
	ServiceInstanceParamsSecretKey = "params"

	// UserProvidedServiceClassName is the class name for user-provided service
	// instances, unless overridden by a mock name.
	// The class name is also used as the label for the service in VCAP_SERVICES.
	UserProvidedServiceClassName = "user-provided"

	// UserProvidedServiceDescription indicates the service is a user provided service.
	UserProvidedServiceDescription = "user-provided"

	// BrokeredServiceDescription indicates the service is managed by a broker.
	BrokeredServiceDescription = "brokered"

	// VolumeServiceDescription indicates the service is managed by a volume broker.
	VolumeServiceDescription = "volume"

	// DefaultServiceInstanceProgressDeadlineSeconds contains the default progress
	// deadline for service instances.
	DefaultServiceInstanceProgressDeadlineSeconds int64 = 30 * 60
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstance is a representation for any type of service instance
// (user-provided or created using a service broker).
type ServiceInstance struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec ServiceInstanceSpec `json:"spec,omitempty"`

	// +optional
	Status ServiceInstanceStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*ServiceInstance)(nil)
var _ apis.Defaultable = (*ServiceInstance)(nil)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceList is a list of ServiceInstance resources
type ServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceInstance `json:"items"`
}

// ServiceInstanceSpec contains the specification for a binding.
type ServiceInstanceSpec struct {
	// ServiceType is a pointer to the type of the service instance.
	ServiceType `json:",inline"`

	// Tags are optional tags provided by the user. They are included in VCAP_SERVICES for a service.
	// Brokered services have tags associated with the CommonServiceClassSpec for that service.
	// The tags set in this field will override all other tags.
	// The JSON encoding of tags in VCAP_SERVICES in Cloud Foundry is [] rather than null, which is why
	// the Tags field is not omitempty.
	Tags []string `json:"tags"`

	// ParametersFrom contains a reference to a secret containing parameters
	// for the service.
	ParametersFrom corev1.LocalObjectReference `json:"parametersFrom,omitempty"`

	// DeleteRequests is a unique identifier for an ServiceInstanceSpec.
	// Updating sub-values will trigger an additional delete retry.
	DeleteRequests int `json:"deleteRequests,omitempty"`
}

// ServiceType is the type of the service instance.
type ServiceType struct {
	// One and only one of the following should be specified.
	// UPS is a user-provided service instance.
	// +optional
	UPS *UPSInstance `json:"userProvided,omitempty"`

	// Brokered is a service instance created using a service broker via KSC.
	// +optional
	Brokered *BrokeredInstance `json:"brokered,omitempty"`

	// Volume is a volume service instance created using volume broker.
	// +optional
	Volume *OSBInstance `json:"volume,omitempty"`

	// OSB is a service instance created using Kf's built-in OSB support.
	// +optional
	OSB *OSBInstance `json:"osb,omitempty"`
}

// UPSInstance is a user-provided service instance.
type UPSInstance struct {
	// RouteServiceURL is an alias for the net/url parsing of the service URL.
	// It is not empty if the service instance is a route service.
	// +optional
	RouteServiceURL *RouteServiceURL `json:"routeServiceURL,omitempty"`

	// MockClassName mocks the name of a different service class.
	// This allows overriding the name as it shows up in VCAP_SERVICES
	// to something other than the default "user-provided".
	MockClassName string `json:"mockClassName,omitempty"`

	// MockPlanName mocks the name of a different plan.
	// This allows overriding the name as it shows up in VCAP_SERVICES
	// to something other than the default blank string.
	MockPlanName string `json:"mockPlanName,omitempty"`
}

// RouteServiceDestination includes the fields for a route service destination as well as the name of the route service.
type RouteServiceDestination struct {
	// Name is the name of the route service instance. For user-provided services, this is the name defined by the user upon creation.
	Name string `json:"name,omitempty"`

	// RouteServiceURL is an alias for the net/url parsing of the service URL.
	RouteServiceURL *RouteServiceURL `json:"routeServiceURL,omitempty"`
}

// BrokeredInstance is a service instance created by a service broker via KSC.
// deprecated
type BrokeredInstance struct {
	// Broker is the name of the service broker for the service instance. Fill this in to explicitly specify a broker
	// if a service class and plan could match to multiple brokers.
	// +optional
	Broker string `json:"broker,omitempty"`

	// ClassName is the name of the service class.
	ClassName string `json:"className,omitempty"`

	// PlanName is the name of the service plan.
	PlanName string `json:"planName,omitempty"`

	// Namespaced is true if the service broker/class/plan is namespaced,
	// and false they are available at the cluster level.
	Namespaced bool `json:"namespaced,omitempty"`
}

// OSBInstance is a service instance created using Kf's built-in OSB support.
type OSBInstance struct {
	// BrokerName is the name of the service broker for the service instance.
	BrokerName string `json:"brokerName,omitempty"`

	// Namespaced is true if the service broker/class/plan is namespaced,
	// and false they are available at the cluster level.
	Namespaced bool `json:"namespaced,omitempty"`

	// ClassUID contains the UID of the class, used for provisioning purposes.
	ClassUID string `json:"classUID,omitempty"`

	// ClassName contains the human-readable name of the class.
	ClassName string `json:"className,omitempty"`

	// PlanUID contains the UID of the plan, used for provisioning purposes.
	PlanUID string `json:"planUID,omitempty"`

	// PlanName contains the human-readable name of the plan.
	PlanName string `json:"planName,omitempty"`

	// ProgressDeadlineSeconds contains a configurable timeout between state
	// transition and reaching a stable state before provisioning or deprovisioning
	// times out.
	ProgressDeadlineSeconds int64 `json:"progressDeadlineSeconds,omitempty"`
}

// ServiceInstanceStatus represents information about the status of a ServiceInstance.
type ServiceInstanceStatus struct {
	// Pull in fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// ServiceTypeDescription is a human-readable name for the type of
	// service referenced by this instance.
	ServiceTypeDescription string `json:"serviceTypeDescription,omitempty"`

	// SecretName is the K8s secret name that stores the parameters for a service instance.
	SecretName string `json:"secretName,omitempty"`

	ServiceFields `json:",inline"`

	// RouteServiceURL is an alias for the net/url parsing of the service URL.
	RouteServiceURL *RouteServiceURL `json:"routeServiceURL,omitempty"`

	// OSBStatus contains information about the lifecycle of the OSB backed
	// service.
	OSBStatus OSBStatus `json:"osbStatus,omitempty"`

	// VolumeStatus contains information about the k8s Volume objects
	VolumeStatus *VolumeStatus `json:"volumeStatus,omitempty"`

	// DeleteRequests is the last processed DeleteRequests value
	DeleteRequests int `json:"deleteRequests,omitempty"`
}

// VolumeStatus is a union of status information for an volume instance.
type VolumeStatus struct {
	// PersistentVolumeName is the name of the PersistentVolume created for
	// this instance.
	PersistentVolumeName string `json:"PersistentVolumeName,omitempty"`

	// PersistentVolumeClaimName is the name of the PersistentVolumeClaim
	// created for this instance.
	PersistentVolumeClaimName string `json:"PersistentVolumeClaimName,omitempty"`
}

// OSBStatus is a union of status information for the state of a particular
// resource. Exactly one should be set at any one time.
type OSBStatus struct {
	Provisioning      *OSBState `json:"provisioning,omitempty"`
	Provisioned       *OSBState `json:"provisioned,omitempty"`
	ProvisionFailed   *OSBState `json:"provisionFailed,omitempty"`
	Deprovisioning    *OSBState `json:"deprovisioning,omitempty"`
	Deprovisioned     *OSBState `json:"deprovisioned,omitempty"`
	DeprovisionFailed *OSBState `json:"deprovisionFailed,omitempty"`
}

// IsBlank returns true if the status is unset.
func (o *OSBStatus) IsBlank() bool {
	return *o == OSBStatus{}
}

// OSBState contains information about a specific state.
type OSBState struct {
	// OperationKey, if specified, holds the long running operation key for a given
	// state. OSB uses this arbitrary value to reference specific back-end tasks
	// it's performing.
	OperationKey *string `json:"operationKey,omitempty"`
}

// ServiceFields are fields related to the service used in VCAP_SERVICES.
type ServiceFields struct {
	// Tags contains a list of tags to apply to the service when injecting
	// via VCAP_SERVICES.
	// The JSON encoding of tags in VCAP_SERVICES in Cloud Foundry is [] rather than null, which is why
	// the Tags field is not omitempty.
	Tags []string `json:"tags"`

	// ClassName contains the human-readable name of the class
	ClassName string `json:"className,omitempty"`

	// PlanName contains the human-readable name of the plan
	PlanName string `json:"planName,omitempty"`
}

// VolumeInstanceParams are the volume related fields stored in the instance secret.
type VolumeInstanceParams struct {
	// Share is the NFS share address.
	Share string `json:"share,omitempty"`

	// Capacity is the requested capacity for this volume instance.
	// Must be parsable to k8s quantity. https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go
	// +optional
	Capacity string `json:"capacity,omitempty"`

	// Version specifies the NFS version to use.
	// +optional
	Version string `json:"version,omitempty"`
}

// ParseVolumeInstanceParams parses VolumeInstanceParams from the secret of the serviceinstance.
func ParseVolumeInstanceParams(secret *corev1.Secret) (*VolumeInstanceParams, error) {
	paramsJSON, ok := secret.Data[ServiceInstanceParamsSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret is missing key %q", ServiceInstanceParamsSecretKey)
	}

	params := VolumeInstanceParams{}
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %v", err)
	}

	return &params, nil
}

// IsLegacyBrokered returns whether the service instance is created with a KSC service broker.
func (service *ServiceInstance) IsLegacyBrokered() bool {
	return service.Spec.Brokered != nil
}

// IsKfBrokered returns whether the service instance is created with a service broker.
func (service *ServiceInstance) IsKfBrokered() bool {
	return service.Spec.OSB != nil
}

// IsUserProvided returns whether the service instance is a user-provided service.
func (service *ServiceInstance) IsUserProvided() bool {
	return service.Spec.UPS != nil
}

// IsRouteService returns whether the service instance is a route service.
// Only user-provided services can be route services at this time.
func (service *ServiceInstance) IsRouteService() bool {
	return service.IsUserProvided() && service.Spec.UPS.RouteServiceURL != nil
}

// IsVolume returns whether the service instance is a volume service.
func (service *ServiceInstance) IsVolume() bool {
	return service.Spec.Volume != nil
}

// HasNoBackingResources returns whether the service instance is a ServiceType that has backing resources.
func (service *ServiceInstance) HasNoBackingResources() bool {
	return service.IsUserProvided() || service.IsVolume()
}
