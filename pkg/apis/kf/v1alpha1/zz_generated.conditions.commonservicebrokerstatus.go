// Copyright 2023 Google LLC
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

	// CommonServiceBrokerConditionReady is set when the CRD is configured and is usable.
	CommonServiceBrokerConditionReady = apis.ConditionReady

	// CommonServiceBrokerConditionCredsSecretReady is set when the child
	// resource(s) CredsSecret is/are ready.
	CommonServiceBrokerConditionCredsSecretReady apis.ConditionType = "CredsSecretReady"

	// CommonServiceBrokerConditionCredsSecretPopulatedReady is set when the child
	// resource(s) CredsSecretPopulated is/are ready.
	CommonServiceBrokerConditionCredsSecretPopulatedReady apis.ConditionType = "CredsSecretPopulatedReady"

	// CommonServiceBrokerConditionCatalogReady is set when the child
	// resource(s) Catalog is/are ready.
	CommonServiceBrokerConditionCatalogReady apis.ConditionType = "CatalogReady"
)

func (status *CommonServiceBrokerStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		CommonServiceBrokerConditionCredsSecretReady,
		CommonServiceBrokerConditionCredsSecretPopulatedReady,
		CommonServiceBrokerConditionCatalogReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *CommonServiceBrokerStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *CommonServiceBrokerStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(CommonServiceBrokerConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *CommonServiceBrokerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *CommonServiceBrokerStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// CredsSecretCondition gets a manager for the state of the child resource.
func (status *CommonServiceBrokerStatus) CredsSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), CommonServiceBrokerConditionCredsSecretReady, "CredsSecret")
}

// CredsSecretPopulatedCondition gets a manager for the state of the child resource.
func (status *CommonServiceBrokerStatus) CredsSecretPopulatedCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), CommonServiceBrokerConditionCredsSecretPopulatedReady, "CredsSecretPopulated")
}

// CatalogCondition gets a manager for the state of the child resource.
func (status *CommonServiceBrokerStatus) CatalogCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), CommonServiceBrokerConditionCatalogReady, "Catalog")
}

func (status *CommonServiceBrokerStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
