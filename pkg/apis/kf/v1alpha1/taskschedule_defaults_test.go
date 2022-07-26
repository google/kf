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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestTaskSchedule_SetDefaults_scheduleAlreadySet(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{
		Spec: TaskScheduleSpec{
			Schedule: "1 * * * *",
		},
	}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "schedule", "1 * * * *", ts.Spec.Schedule)
}

func TestTaskSchedule_SetDefaults_schedule(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "schedule", "* * * * *", ts.Spec.Schedule)
}

func TestTaskSchedule_SetDefaults_concurrencyPolicyAlreadySet(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{
		Spec: TaskScheduleSpec{
			ConcurrencyPolicy: "Forbid",
		},
	}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "concurrencyPolicy", "Forbid", ts.Spec.ConcurrencyPolicy)
}

func TestTaskSchedule_SetDefaults_concurrencyPolicy(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "concurrencyPolicy", "Always", ts.Spec.ConcurrencyPolicy)
}

func TestTaskSchedule_SetDefaults_labels(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "managedBy", "kf", ts.Labels[ManagedByLabel])
	testutil.AssertEqual(t, "component", "task-schedule", ts.Labels[ComponentLabel])
	testutil.AssertEqual(t, "suspend", "false", ts.Labels[TaskScheduleSuspendLabel])
}

func TestTaskSchedule_SetDefaults_suspended(t *testing.T) {
	t.Parallel()

	ts := &TaskSchedule{
		Spec: TaskScheduleSpec{
			Suspend: true,
		},
	}
	ts.SetDefaults(context.Background())

	testutil.AssertEqual(t, "suspend", "true", ts.Labels[TaskScheduleSuspendLabel])
}
