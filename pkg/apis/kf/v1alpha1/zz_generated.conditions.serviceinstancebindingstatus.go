// Copyright 2022 Google LLC
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

// This file was generated with conditiongen/generator.go, DO NOT EDIT IT.

package v1alpha1

import (
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// ConditionType represents a Service condition value
const (

	// ServiceInstanceBindingConditionReady is set when the CRD is configured and is usable.
	ServiceInstanceBindingConditionReady = apis.ConditionReady

	// ServiceInstanceBindingConditionServiceInstanceReady is set when the child
	// resource(s) ServiceInstance is/are ready.
	ServiceInstanceBindingConditionServiceInstanceReady apis.ConditionType = "ServiceInstanceReady"

	// ServiceInstanceBindingConditionBackingResourceReady is set when the child
	// resource(s) BackingResource is/are ready.
	ServiceInstanceBindingConditionBackingResourceReady apis.ConditionType = "BackingResourceReady"

	// ServiceInstanceBindingConditionParamsSecretReady is set when the child
	// resource(s) ParamsSecret is/are ready.
	ServiceInstanceBindingConditionParamsSecretReady apis.ConditionType = "ParamsSecretReady"

	// ServiceInstanceBindingConditionParamsSecretPopulatedReady is set when the child
	// resource(s) ParamsSecretPopulated is/are ready.
	ServiceInstanceBindingConditionParamsSecretPopulatedReady apis.ConditionType = "ParamsSecretPopulatedReady"

	// ServiceInstanceBindingConditionCredentialsSecretReady is set when the child
	// resource(s) CredentialsSecret is/are ready.
	ServiceInstanceBindingConditionCredentialsSecretReady apis.ConditionType = "CredentialsSecretReady"

	// ServiceInstanceBindingConditionVolumeParamsPopulatedReady is set when the child
	// resource(s) VolumeParamsPopulated is/are ready.
	ServiceInstanceBindingConditionVolumeParamsPopulatedReady apis.ConditionType = "VolumeParamsPopulatedReady"
)

func (status *ServiceInstanceBindingStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		ServiceInstanceBindingConditionServiceInstanceReady,
		ServiceInstanceBindingConditionBackingResourceReady,
		ServiceInstanceBindingConditionParamsSecretReady,
		ServiceInstanceBindingConditionParamsSecretPopulatedReady,
		ServiceInstanceBindingConditionCredentialsSecretReady,
		ServiceInstanceBindingConditionVolumeParamsPopulatedReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *ServiceInstanceBindingStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *ServiceInstanceBindingStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(ServiceInstanceBindingConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *ServiceInstanceBindingStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *ServiceInstanceBindingStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// ServiceInstanceCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) ServiceInstanceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionServiceInstanceReady, "ServiceInstance")
}

// BackingResourceCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) BackingResourceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionBackingResourceReady, "BackingResource")
}

// ParamsSecretCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) ParamsSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionParamsSecretReady, "ParamsSecret")
}

// ParamsSecretPopulatedCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) ParamsSecretPopulatedCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionParamsSecretPopulatedReady, "ParamsSecretPopulated")
}

// CredentialsSecretCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) CredentialsSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionCredentialsSecretReady, "CredentialsSecret")
}

// VolumeParamsPopulatedCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceBindingStatus) VolumeParamsPopulatedCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceBindingConditionVolumeParamsPopulatedReady, "VolumeParamsPopulated")
}

func (status *ServiceInstanceBindingStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
