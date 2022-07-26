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
	"strconv"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// ServiceInstanceBindingParamsSecretKey contains the secret key that holds parameters.
	ServiceInstanceBindingParamsSecretKey = "params"

	// DefaultServiceInstanceBindingProgressDeadlineSeconds contains the default
	// amount of time bindings can take before timing out.
	DefaultServiceInstanceBindingProgressDeadlineSeconds = DefaultServiceInstanceProgressDeadlineSeconds
)

// MakeServiceBindingName returns a deterministic name for a service instance binding.
func MakeServiceBindingName(appName, instanceName string) string {
	return GenerateName("binding", appName, instanceName)
}

// MakeServiceBindingParamsSecretName returns a deterministic name for the binding parameters secret.
func MakeServiceBindingParamsSecretName(appName, instanceName string) string {
	return GenerateName("binding", appName, instanceName, "params")
}

// MakeRouteServiceBindingName returns a deterministic name for a Route service instance binding.
func MakeRouteServiceBindingName(hostname, domain, path, instanceName string) string {
	rsf := RouteSpecFields{
		Hostname: hostname,
		Domain:   domain,
		Path:     path,
	}
	return GenerateName("binding", rsf.String(), instanceName)
}

// MakeRouteServiceBindingParamsSecretName returns a deterministic name for a Route binding parameters secret.
func MakeRouteServiceBindingParamsSecretName(hostname, domain, path, instanceName string) string {
	rsf := RouteSpecFields{
		Hostname: hostname,
		Domain:   domain,
		Path:     path,
	}
	return GenerateName("binding", rsf.String(), instanceName, "params")
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceBinding is an abstraction for a service binding between any type of ServiceInstance and an App.
type ServiceInstanceBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec ServiceInstanceBindingSpec `json:"spec,omitempty"`

	// +optional
	Status ServiceInstanceBindingStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*ServiceInstanceBinding)(nil)
var _ apis.Defaultable = (*ServiceInstanceBinding)(nil)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceBindingList is a list of Binding resources
type ServiceInstanceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceInstanceBinding `json:"items"`
}

// ServiceInstanceBindingSpec contains the specification for a binding.
type ServiceInstanceBindingSpec struct {
	// BindingType is a pointer to the type of the service instance binding.
	BindingType `json:",inline"`

	// InstanceRef is the service instance that is bound to the App or Route.
	InstanceRef core.LocalObjectReference `json:"instanceRef"`

	// ParametersFrom contains a reference to a secret containing parameters for
	// the service instance binding.
	ParametersFrom core.LocalObjectReference `json:"parametersFrom,omitempty"`

	// BindingNameOverride is the custom binding name set by the user. If it is not set, the name of the service instance is used.
	// +optional
	BindingNameOverride string `json:"bindingNameOverride,omitempty"`

	// ProgressDeadlineSeconds contains a configurable timeout between state
	// transition and reaching a stable state before binding or unbinding
	// times out.
	ProgressDeadlineSeconds int64 `json:"progressDeadlineSeconds,omitempty"`

	// UnbindRequests is a unique identifier for an ServiceInstanceBindingSpec.
	// Updating sub-values will trigger an additional unbind retry.
	UnbindRequests int `json:"unbindRequests,omitempty"`
}

// BindingType is the type of the service instance binding.
type BindingType struct {
	// One and only one of the following binding types should be specified.

	// App is the Kf App that the service instance is bound to.
	// +optional
	App *AppRef `json:"app,omitempty"`

	// Route is the Route that the service instance is bound to.
	// +optional
	Route *RouteRef `json:"route,omitempty"`
}

type AppRef core.LocalObjectReference

type RouteRef RouteSpecFields

// ServiceInstanceBindingStatus represents information about the status of a Binding.
type ServiceInstanceBindingStatus struct {
	// Pull in fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// BindingName is the custom binding name set by the user, or the name of the service instance if a custom name was not provided.
	BindingName string `json:"bindingName,omitempty"`

	// CredentialsSecretRef is the K8s secret name that stores the credentials for the service binding.
	CredentialsSecretRef core.LocalObjectReference `json:"credentialsSecretRef,omitempty"`

	// ServiceFields is the set of fields related to the service instance in the binding, used for VCAP_SERVICES.
	ServiceFields `json:",inline"`

	// VolumeStatus contains information about the k8s Volume objects
	VolumeStatus *BindingVolumeStatus `json:"volumeStatus,omitempty"`

	// RouteServiceURL is an alias for the net/url parsing of the service URL.
	RouteServiceURL *RouteServiceURL `json:"routeServiceURL,omitempty"`

	// OSBStatus contains information about the lifecycle of the OSB backed
	// service.
	OSBStatus BindingOSBStatus `json:"osbStatus,omitempty"`

	// UnbindRequests is the last processed UnbindRequests value
	UnbindRequests int `json:"unbindRequests,omitempty"`
}

// BindingVolumeParams are the volume related fields stored in the binding's secret.
type BindingVolumeParams struct {
	// Mount is the path to mount the NFS share.
	Mount string `json:"mount"`

	// ReadOnly indicates whether the mounted share is readonly.
	ReadOnly bool `json:"readonly,omitempty"`

	UidGid `json:",inline"`
}

// UidGid contains the UID and GID fields and the necessary unmarshal logic.
type UidGid struct {

	// UID if specified will change the owner of the mounted directory to UID.
	// +kubebuilder:validation:Type=string
	// +nullable
	// +optional
	UID ID `json:"UID,omitempty" type:"string"`

	// GID if specified will change the group of the mounted directory to GID.
	// +kubebuilder:validation:Type=string
	// +nullable
	// +optional
	GID ID `json:"GID,omitempty" type:"string"`
}

// ID is an Int64 that is very flexible. If a string is passed in instead, it
// will assume a 0 value.
// NOTE: This has to be a custom type for the UnmarshalJSON (as opposed to
// just using a string with an embed type) so that the JSON unmarshaler still
// unmarshals the outer type. See https://github.com/golang/go/issues/39470
// for more info.
type ID string

// UnmarshalJSON implements json.Unmarshaler. This is necessary as the UID and
// GID can be a bit tricky with strings vs integers. We *wanted* them to be
// strings, but they can look like ints which can throw off the parsers.
func (g *ID) UnmarshalJSON(data []byte) error {
	var i interface{}
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	switch x := i.(type) {
	case string:
		*g = ID(x)
	case int, int32, int64, uint, uint32, uint64:
		*g = ID(fmt.Sprint(x))
	default:
		*g = ""
	}

	return nil
}

// UIDInt64 returns the UID as an Int64.
func (g *UidGid) UIDInt64() (int64, error) {
	if uid, err := strconv.ParseInt(string(g.UID), 10, 64); err != nil {
		return 0, err
	} else if uid < 0 {
		return 0, fmt.Errorf("must be greater than or equal to 0")
	} else {
		return uid, nil
	}
}

// GIDInt64 returns the GID as an Int64.
func (g *UidGid) GIDInt64() (int64, error) {
	if gid, err := strconv.ParseInt(string(g.GID), 10, 64); err != nil {
		return 0, err
	} else if gid < 0 {
		return 0, fmt.Errorf("must be greater than or equal to 0")
	} else {
		return gid, nil
	}
}

// BindingVolumeStatus is the volume related status.
type BindingVolumeStatus struct {
	// Mount is the path to mount the NFS share.
	Mount string `json:"mount"`

	// PersistentVolumeName is the name of the binded PersistentVolume.
	PersistentVolumeName string `json:"volumeName,omitempty"`

	// PersistentVolumeClaimName is the name of the binded PersistentVolumeClaim.
	PersistentVolumeClaimName string `json:"claimName,omitempty"`

	// ReadOnly indicates whether the mounted share is readonly.
	ReadOnly bool `json:"readonly,omitempty"`

	UidGid `json:",inline"`
}

// BindingOSBStatus is a union of status information for the state of a
// particular binding. Exactly one should be set at any one time.
type BindingOSBStatus struct {
	Binding      *OSBState `json:"binding,omitempty"`
	Bound        *OSBState `json:"bound,omitempty"`
	BindFailed   *OSBState `json:"bindFailed,omitempty"`
	Unbinding    *OSBState `json:"unbinding,omitempty"`
	Unbound      *OSBState `json:"unbound,omitempty"`
	UnbindFailed *OSBState `json:"unbindFailed,omitempty"`
}

// IsBlank returns true if the status is unset.
func (o *BindingOSBStatus) IsBlank() bool {
	return *o == BindingOSBStatus{}
}

// IsAppBinding returns true if the service instance binding binds a service to an App.
func (binding *ServiceInstanceBinding) IsAppBinding() bool {
	return binding.Spec.BindingType.App != nil
}

// IsRouteBinding returns true if the service instance binding binds a service to a Route.
func (binding *ServiceInstanceBinding) IsRouteBinding() bool {
	return binding.Spec.BindingType.Route != nil
}
