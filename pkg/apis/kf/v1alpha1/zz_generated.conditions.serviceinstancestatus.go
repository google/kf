// Copyright 2025 Google LLC
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

	// ServiceInstanceConditionReady is set when the CRD is configured and is usable.
	ServiceInstanceConditionReady = apis.ConditionReady

	// ServiceInstanceConditionSpaceReady is set when the child
	// resource(s) Space is/are ready.
	ServiceInstanceConditionSpaceReady apis.ConditionType = "SpaceReady"

	// ServiceInstanceConditionBackingResourceReady is set when the child
	// resource(s) BackingResource is/are ready.
	ServiceInstanceConditionBackingResourceReady apis.ConditionType = "BackingResourceReady"

	// ServiceInstanceConditionParamsSecretReady is set when the child
	// resource(s) ParamsSecret is/are ready.
	ServiceInstanceConditionParamsSecretReady apis.ConditionType = "ParamsSecretReady"

	// ServiceInstanceConditionParamsSecretPopulatedReady is set when the child
	// resource(s) ParamsSecretPopulated is/are ready.
	ServiceInstanceConditionParamsSecretPopulatedReady apis.ConditionType = "ParamsSecretPopulatedReady"
)

func (status *ServiceInstanceStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		ServiceInstanceConditionSpaceReady,
		ServiceInstanceConditionBackingResourceReady,
		ServiceInstanceConditionParamsSecretReady,
		ServiceInstanceConditionParamsSecretPopulatedReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *ServiceInstanceStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *ServiceInstanceStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(ServiceInstanceConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *ServiceInstanceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *ServiceInstanceStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// SpaceCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceStatus) SpaceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceConditionSpaceReady, "Space")
}

// BackingResourceCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceStatus) BackingResourceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceConditionBackingResourceReady, "BackingResource")
}

// ParamsSecretCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceStatus) ParamsSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceConditionParamsSecretReady, "ParamsSecret")
}

// ParamsSecretPopulatedCondition gets a manager for the state of the child resource.
func (status *ServiceInstanceStatus) ParamsSecretPopulatedCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), ServiceInstanceConditionParamsSecretPopulatedReady, "ParamsSecretPopulated")
}

func (status *ServiceInstanceStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
