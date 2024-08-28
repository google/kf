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

	// SourcePackageConditionSucceeded is set when the CRD is completed.
	SourcePackageConditionSucceeded = apis.ConditionSucceeded

	// SourcePackageConditionUploadReady is set when the child
	// resource(s) Upload is/are ready.
	SourcePackageConditionUploadReady apis.ConditionType = "UploadReady"
)

func (status *SourcePackageStatus) manage() apis.ConditionManager {
	return apis.NewBatchConditionSet(
		SourcePackageConditionUploadReady,
	).Manage(status)
}

// Succeeded returns if the type successfully completed.
func (status *SourcePackageStatus) Succeeded() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *SourcePackageStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *SourcePackageStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// UploadCondition gets a manager for the state of the child resource.
func (status *SourcePackageStatus) UploadCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SourcePackageConditionUploadReady, "Upload")
}

func (status *SourcePackageStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
