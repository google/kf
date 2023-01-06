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

	// TaskConditionSucceeded is set when the CRD is completed.
	TaskConditionSucceeded = apis.ConditionSucceeded

	// TaskConditionSpaceReady is set when the child
	// resource(s) Space is/are ready.
	TaskConditionSpaceReady apis.ConditionType = "SpaceReady"

	// TaskConditionPipelineRunReady is set when the child
	// resource(s) PipelineRun is/are ready.
	TaskConditionPipelineRunReady apis.ConditionType = "PipelineRunReady"

	// TaskConditionTaskRunReady is set when the child
	// resource(s) TaskRun is/are ready.
	TaskConditionTaskRunReady apis.ConditionType = "TaskRunReady"

	// TaskConditionAppReady is set when the child
	// resource(s) App is/are ready.
	TaskConditionAppReady apis.ConditionType = "AppReady"

	// TaskConditionConfigReady is set when the child
	// resource(s) Config is/are ready.
	TaskConditionConfigReady apis.ConditionType = "ConfigReady"
)

func (status *TaskStatus) manage() apis.ConditionManager {
	return apis.NewBatchConditionSet(
		TaskConditionSpaceReady,
		TaskConditionPipelineRunReady,
		TaskConditionTaskRunReady,
		TaskConditionAppReady,
		TaskConditionConfigReady,
	).Manage(status)
}

// Succeeded returns if the type successfully completed.
func (status *TaskStatus) Succeeded() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *TaskStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *TaskStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// SpaceCondition gets a manager for the state of the child resource.
func (status *TaskStatus) SpaceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), TaskConditionSpaceReady, "Space")
}

// PipelineRunCondition gets a manager for the state of the child resource.
func (status *TaskStatus) PipelineRunCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), TaskConditionPipelineRunReady, "PipelineRun")
}

// TaskRunCondition gets a manager for the state of the child resource.
func (status *TaskStatus) TaskRunCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), TaskConditionTaskRunReady, "TaskRun")
}

// AppCondition gets a manager for the state of the child resource.
func (status *TaskStatus) AppCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), TaskConditionAppReady, "App")
}

// ConfigCondition gets a manager for the state of the child resource.
func (status *TaskStatus) ConfigCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), TaskConditionConfigReady, "Config")
}

func (status *TaskStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
