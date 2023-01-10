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

	// BuildConditionSucceeded is set when the CRD is completed.
	BuildConditionSucceeded = apis.ConditionSucceeded

	// BuildConditionSpaceReady is set when the child
	// resource(s) Space is/are ready.
	BuildConditionSpaceReady apis.ConditionType = "SpaceReady"

	// BuildConditionTaskRunReady is set when the child
	// resource(s) TaskRun is/are ready.
	BuildConditionTaskRunReady apis.ConditionType = "TaskRunReady"

	// BuildConditionSourcePackageReady is set when the child
	// resource(s) SourcePackage is/are ready.
	BuildConditionSourcePackageReady apis.ConditionType = "SourcePackageReady"
)

func (status *BuildStatus) manage() apis.ConditionManager {
	return apis.NewBatchConditionSet(
		BuildConditionSpaceReady,
		BuildConditionTaskRunReady,
		BuildConditionSourcePackageReady,
	).Manage(status)
}

// Succeeded returns if the type successfully completed.
func (status *BuildStatus) Succeeded() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *BuildStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *BuildStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// SpaceCondition gets a manager for the state of the child resource.
func (status *BuildStatus) SpaceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), BuildConditionSpaceReady, "Space")
}

// TaskRunCondition gets a manager for the state of the child resource.
func (status *BuildStatus) TaskRunCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), BuildConditionTaskRunReady, "TaskRun")
}

// SourcePackageCondition gets a manager for the state of the child resource.
func (status *BuildStatus) SourcePackageCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), BuildConditionSourcePackageReady, "SourcePackage")
}

func (status *BuildStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
