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
	"context"
	"strconv"
)

const (
	TaskScheduleSuspendLabel  = "taskschedules.kf.dev/suspend"
	taskScheduleComponentName = "task-schedule"
)

// SetDefaults implements apis.Defaultable.
func (k *TaskSchedule) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
	k.Labels = UnionMaps(
		k.Labels,
		map[string]string{
			ManagedByLabel: "kf",
			ComponentLabel: taskScheduleComponentName,
			// TODO(https://github.com/kubernetes/kubernetes/issues/53459)
			// We need to filter based on the value of spec.suspend; however,
			// CRDs do not currently support field selectors. Copy the field
			// value over to a label and use label selectors until k8s adds
			// field selector support to CRDs.
			TaskScheduleSuspendLabel: strconv.FormatBool(k.Spec.Suspend),
		},
	)
}

const (
	defaultSchedule = "* * * * *"
)

// SetDefaults implements apis.Defaultable.
func (k *TaskScheduleSpec) SetDefaults(ctx context.Context) {
	if k.Schedule == "" {
		k.Schedule = defaultSchedule
	}

	if k.ConcurrencyPolicy == "" {
		k.ConcurrencyPolicy = ConcurrencyPolicyAlways
	}
}
