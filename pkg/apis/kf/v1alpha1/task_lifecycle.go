// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (t *Task) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Task")
}

// PropagateTaskStatus copies fields from the Tekton TaskRun and
// updates the Task readiness based on the TaskRun state.
func (status *TaskStatus) PropagateTaskStatus(tr *tektonv1beta1.TaskRun) {
	if tr == nil {
		return
	}

	status.StartTime = tr.Status.StartTime
	status.CompletionTime = tr.Status.CompletionTime

	// Generate a task duration for easy printing.
	if status.StartTime != nil && status.CompletionTime != nil {
		status.Duration = &metav1.Duration{
			Duration: status.CompletionTime.Time.Sub(status.StartTime.Time),
		}
	}

	cond := tr.Status.GetCondition(apis.ConditionSucceeded)
	PropagateCondition(status.manage(), TaskConditionTaskRunReady, cond)
}

// PropagateTerminatingStatus updates the ready status of the Task to False if the Task received a delete request.
func (status *TaskStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(TaskConditionSucceeded, "Terminating", "Task is terminating")
}

// MarkSpaceHealthy notes that the Space was able to be retrieved and
// defaults can be applied from it.
func (status *TaskStatus) MarkSpaceHealthy() {
	status.SpaceCondition().MarkSuccess()
}

// MarkSpaceUnhealthy notes that the Space was could not be retrieved.
func (status *TaskStatus) MarkSpaceUnhealthy(reason, message string) {
	status.SpaceCondition().MarkFalse(reason, message)
}
