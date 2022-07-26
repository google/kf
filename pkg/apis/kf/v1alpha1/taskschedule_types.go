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
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// Runs scheduled jobs regardless of status of previously run jobs.
	ConcurrencyPolicyAlways = "Always"
	// Skips scheduling new jobs while a previous execution is still running.
	ConcurrencyPolicyForbid = "Forbid"
	// Cancels any still running jobs from the same schedule when a new job is started.
	ConcurrencyPolicyReplace = "Replace"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskSchedule is a configuration to create Tasks on a cron schedule.
type TaskSchedule struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec TaskScheduleSpec `json:"spec,omitempty"`

	// +optional
	Status TaskScheduleStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*TaskSchedule)(nil)
var _ apis.Defaultable = (*TaskSchedule)(nil)

// TaskScheduleSpec contains the specification of a TaskSchedule.
type TaskScheduleSpec struct {
	// ConcurrencyPolicy specifies how to treat concurrent executions of Tasks.
	// Valid values are
	//
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run
	// 		hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one.
	ConcurrencyPolicy string `json:"concurrencyPolicy,omitempty"`

	// Schedule is the interval to start Tasks in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// Suspend tells the controller to suspend subsequent executions. It does
	// not apply to already started executions.
	// +optional
	Suspend bool `json:"suspend"`

	// TaskTemplate specifies the Task that will be created when executing a TaskSchedule.
	TaskTemplate TaskSpec `json:"taskTemplate,omitempty"`
}

// TaskScheduleStatus represents information about the status of a TaskSchedule.
type TaskScheduleStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	TaskScheduleStatusFields `json:",inline"`
}

// TaskScheduleStatusFields hold the fields of Task's status that
// are shared.
type TaskScheduleStatusFields struct {
	// Active is a list of currently running Tasks created via this TaskSchedule.
	Active []corev1.LocalObjectReference `json:"active,omitempty"`

	// LastScheduleTime is the timestamp of when a Task was last scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// TODO(b/193059618): Consider adding additional Status fields (enumerated in bug).
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskScheduleList is a list of TaskSchedule resources.
type TaskScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TaskSchedule `json:"items"`
}
