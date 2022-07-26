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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// TaskComponentName holds the component label anme for Task.
	TaskComponentName = "task"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (t *TaskSchedule) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("TaskSchedule")
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Task is a representation for short-lived task.
type Task struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec TaskSpec `json:"spec,omitempty"`

	// +optional
	Status TaskStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*Task)(nil)
var _ apis.Defaultable = (*Task)(nil)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList is a list of Task resources.
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Task `json:"items"`
}

// TaskSpec contains the specification of a Task.
type TaskSpec struct {

	// DisplayName of the Task, it is either user-provided or auto generated.
	DisplayName string `json:"displayName,omitempty"`

	// AppRef is to reference the App the task is created on.
	AppRef corev1.LocalObjectReference `json:"appRef,omitempty"`

	// CPU is the number of cpu core to request for the Task, e.g. "1", "500m" or "0.5".
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory is the number of memory units to request for the Task, e.g. "1G", "2Gi".
	// +optional
	Memory string `json:"memory,omitempty"`

	// Disk is the number of ephermeral storage units to request for the Task, e.g. "1G", "2Gi".
	// +optional
	Disk string `json:"disk,omitempty"`

	// Command is the start command to be set for the Task.
	// +optional
	Command string `json:"command,omitempty"`

	// Terminated determines if the Task should have been terminated or not.
	// +optional
	Terminated bool `json:"terminated,omitempty"`
}

// TaskStatus represents information about the status of a Task.
type TaskStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	TaskStatusFields `json:",inline"`
}

// TaskStatusFields hold the fields of Task's status that
// are shared.
type TaskStatusFields struct {
	// ID is a unique identifier of the Task within an App.
	ID int `json:"id,omitempty"`

	// StartTime is the timestamp of when the Task starts.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the timestamp of when the Task completes.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Duration is the time duration of how long did it take for the
	// Task to transition from start to completion.
	Duration *metav1.Duration `json:"duration,omitempty"`
}
