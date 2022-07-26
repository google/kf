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

package taskschedules_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fakeclient "github.com/google/kf/v2/pkg/client/kf/injection/client/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/taskschedules"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteJobSchedule(t *testing.T) {
	t.Parallel()
	const (
		spaceName = "my-space"
		appName   = "my-app"
		jobName   = "my-job"
	)
	var (
		taskSchedule = &v1alpha1.TaskSchedule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: spaceName,
			},
			Spec: v1alpha1.TaskScheduleSpec{
				Schedule: "* * * * *",
			},
		}
		suspendedTaskSchedule = &v1alpha1.TaskSchedule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: spaceName,
			},
			Spec: v1alpha1.TaskScheduleSpec{
				Schedule: "0 0 30 2 *",
				Suspend:  true,
			},
		}
	)
	cases := []struct {
		name      string
		space     string
		args      []string
		setup     func(ctx context.Context, t *testing.T)
		assert    func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error)
		expectErr error
	}{
		{
			name:      "missing args",
			expectErr: errors.New("accepts between 1 and 2 arg(s), received 0"),
		},
		{
			name:      "no target space",
			args:      []string{jobName},
			expectErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		{
			name:      "TaskSchedule does not exist",
			space:     spaceName,
			args:      []string{"non-existent"},
			expectErr: errors.New("taskschedules.kf.dev \"non-existent\" not found"),
		},
		{
			name:  "TaskSchedule is already suspended",
			space: spaceName,
			args:  []string{jobName},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, suspendedTaskSchedule, metav1.CreateOptions{})
			},
			expectErr: errors.New("Job is already suspended."),
		},
		{
			name:  "TaskSchedule is scheduled",
			space: spaceName,
			args:  []string{jobName},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, taskSchedule, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				client := fakeclient.Get(ctx)
				ts, err := client.KfV1alpha1().
					TaskSchedules(spaceName).
					Get(ctx, jobName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "suspend", true, ts.Spec.Suspend)
				testutil.AssertContainsAll(t, buffer.String(), []string{fmt.Sprintf("Job schedule for %s deleted", jobName)})
			},
		},
		{
			name:  "warns on extra arg",
			space: spaceName,
			args:  []string{jobName, "schedule-guid"},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, taskSchedule, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				client := fakeclient.Get(ctx)
				ts, err := client.KfV1alpha1().
					TaskSchedules(spaceName).
					Get(ctx, jobName, metav1.GetOptions{})
				testutil.AssertContainsAll(t, buffer.String(), []string{
					"WARNING: Ignoring SCHEDULE-GUID arg. Kf only supports a single schedule per Job.",
				})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "suspend", true, ts.Spec.Suspend)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gomock.NewController(t)
			cmd := taskschedules.NewDeleteJobScheduleCommand(&config.KfParams{
				Space: tc.space,
			})
			var buffer bytes.Buffer
			ctx := fakeinjection.WithInjection(context.Background(), t)
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)
			cmd.SetOutput(&buffer)
			if tc.setup != nil {
				tc.setup(ctx, t)
			}
			gotErr := cmd.Execute()
			if tc.expectErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, gotErr)
			}
			if tc.assert != nil {
				tc.assert(ctx, t, &buffer, gotErr)
			}
			if gotErr != nil {
				return
			}

		})
	}
}
