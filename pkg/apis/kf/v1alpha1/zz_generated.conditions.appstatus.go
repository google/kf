// Copyright 2024 Google LLC
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

	// AppConditionReady is set when the CRD is configured and is usable.
	AppConditionReady = apis.ConditionReady

	// AppConditionBuildReady is set when the child
	// resource(s) Build is/are ready.
	AppConditionBuildReady apis.ConditionType = "BuildReady"

	// AppConditionServiceReady is set when the child
	// resource(s) Service is/are ready.
	AppConditionServiceReady apis.ConditionType = "ServiceReady"

	// AppConditionServiceAccountReady is set when the child
	// resource(s) ServiceAccount is/are ready.
	AppConditionServiceAccountReady apis.ConditionType = "ServiceAccountReady"

	// AppConditionDeploymentReady is set when the child
	// resource(s) Deployment is/are ready.
	AppConditionDeploymentReady apis.ConditionType = "DeploymentReady"

	// AppConditionSpaceReady is set when the child
	// resource(s) Space is/are ready.
	AppConditionSpaceReady apis.ConditionType = "SpaceReady"

	// AppConditionRouteReady is set when the child
	// resource(s) Route is/are ready.
	AppConditionRouteReady apis.ConditionType = "RouteReady"

	// AppConditionEnvVarSecretReady is set when the child
	// resource(s) EnvVarSecret is/are ready.
	AppConditionEnvVarSecretReady apis.ConditionType = "EnvVarSecretReady"

	// AppConditionServiceInstanceBindingsReady is set when the child
	// resource(s) ServiceInstanceBindings is/are ready.
	AppConditionServiceInstanceBindingsReady apis.ConditionType = "ServiceInstanceBindingsReady"

	// AppConditionHorizontalPodAutoscalerReady is set when the child
	// resource(s) HorizontalPodAutoscaler is/are ready.
	AppConditionHorizontalPodAutoscalerReady apis.ConditionType = "HorizontalPodAutoscalerReady"
)

func (status *AppStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		AppConditionBuildReady,
		AppConditionServiceReady,
		AppConditionServiceAccountReady,
		AppConditionDeploymentReady,
		AppConditionSpaceReady,
		AppConditionRouteReady,
		AppConditionEnvVarSecretReady,
		AppConditionServiceInstanceBindingsReady,
		AppConditionHorizontalPodAutoscalerReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *AppStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *AppStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(AppConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *AppStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *AppStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// BuildCondition gets a manager for the state of the child resource.
func (status *AppStatus) BuildCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionBuildReady, "Build")
}

// ServiceCondition gets a manager for the state of the child resource.
func (status *AppStatus) ServiceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionServiceReady, "Service")
}

// ServiceAccountCondition gets a manager for the state of the child resource.
func (status *AppStatus) ServiceAccountCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionServiceAccountReady, "ServiceAccount")
}

// DeploymentCondition gets a manager for the state of the child resource.
func (status *AppStatus) DeploymentCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionDeploymentReady, "Deployment")
}

// SpaceCondition gets a manager for the state of the child resource.
func (status *AppStatus) SpaceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionSpaceReady, "Space")
}

// RouteCondition gets a manager for the state of the child resource.
func (status *AppStatus) RouteCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionRouteReady, "Route")
}

// EnvVarSecretCondition gets a manager for the state of the child resource.
func (status *AppStatus) EnvVarSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionEnvVarSecretReady, "EnvVarSecret")
}

// ServiceInstanceBindingsCondition gets a manager for the state of the child resource.
func (status *AppStatus) ServiceInstanceBindingsCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionServiceInstanceBindingsReady, "ServiceInstanceBindings")
}

// HorizontalPodAutoscalerCondition gets a manager for the state of the child resource.
func (status *AppStatus) HorizontalPodAutoscalerCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionHorizontalPodAutoscalerReady, "HorizontalPodAutoscaler")
}

func (status *AppStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
