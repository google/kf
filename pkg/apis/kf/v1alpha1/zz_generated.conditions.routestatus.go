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

	// RouteConditionReady is set when the CRD is configured and is usable.
	RouteConditionReady = apis.ConditionReady

	// RouteConditionVirtualServiceReady is set when the child
	// resource(s) VirtualService is/are ready.
	RouteConditionVirtualServiceReady apis.ConditionType = "VirtualServiceReady"

	// RouteConditionSpaceDomainReady is set when the child
	// resource(s) SpaceDomain is/are ready.
	RouteConditionSpaceDomainReady apis.ConditionType = "SpaceDomainReady"

	// RouteConditionRouteServiceReady is set when the child
	// resource(s) RouteService is/are ready.
	RouteConditionRouteServiceReady apis.ConditionType = "RouteServiceReady"
)

func (status *RouteStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		RouteConditionVirtualServiceReady,
		RouteConditionSpaceDomainReady,
		RouteConditionRouteServiceReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *RouteStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *RouteStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(RouteConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *RouteStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *RouteStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// VirtualServiceCondition gets a manager for the state of the child resource.
func (status *RouteStatus) VirtualServiceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), RouteConditionVirtualServiceReady, "VirtualService")
}

// SpaceDomainCondition gets a manager for the state of the child resource.
func (status *RouteStatus) SpaceDomainCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), RouteConditionSpaceDomainReady, "SpaceDomain")
}

// RouteServiceCondition gets a manager for the state of the child resource.
func (status *RouteStatus) RouteServiceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), RouteConditionRouteServiceReady, "RouteService")
}

func (status *RouteStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
