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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fakeclient "github.com/google/kf/v2/pkg/client/kf/injection/client/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/commands/taskschedules"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateJob(t *testing.T) {
	t.Parallel()

	const (
		spaceName = "my-space"
		appName   = "my-app"
		jobName   = "my-job"
		command   = "sleep 100"
	)

	var (
		app = &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: spaceName,
			},
		}

		taskSchedule = &v1alpha1.TaskSchedule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: spaceName,
			},
			Spec: v1alpha1.TaskScheduleSpec{
				Schedule: "* * * * *",
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
			expectErr: errors.New("accepts 3 arg(s), received 0"),
		},
		{
			name:      "no target space",
			args:      []string{appName, jobName, command},
			expectErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		{
			name:      "App does not exist",
			space:     spaceName,
			args:      []string{appName, jobName, command},
			expectErr: errors.New("failed to get App: apps.kf.dev \"my-app\" not found"),
		},
		{
			name:  "TaskSchedule already exists",
			space: spaceName,
			args:  []string{appName, jobName, command},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					Apps(spaceName).
					Create(ctx, app, metav1.CreateOptions{})
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, taskSchedule, metav1.CreateOptions{})
			},
			expectErr: errors.New("failed to create Job: taskschedules.kf.dev \"my-job\" already exists"),
		},
		{
			name:  "create TaskSchedule succeeds",
			space: spaceName,
			args:  []string{appName, jobName, command, "--async"},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					Apps(spaceName).
					Create(ctx, app, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(
					t,
					"output",
					"Job my-job created.\nWaiting for Job to become ready asynchronously\n",
					buffer.String())
				client := fakeclient.Get(ctx)
				ts, err := client.KfV1alpha1().
					TaskSchedules(spaceName).
					Get(context.Background(), jobName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "schedule", "0 0 30 2 *", ts.Spec.Schedule)
				testutil.AssertEqual(t, "suspend", true, ts.Spec.Suspend)
			},
		},
		{
			name:  "sets schedule if provided",
			space: spaceName,
			args:  []string{appName, jobName, command, "--schedule", "* * * * *", "--async"},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					Apps(spaceName).
					Create(ctx, app, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(
					t,
					"output",
					"Job my-job created.\nWaiting for Job to become ready asynchronously\n",
					buffer.String())
				client := fakeclient.Get(ctx)
				ts, err := client.KfV1alpha1().
					TaskSchedules(spaceName).
					Get(context.Background(), jobName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "schedule", "* * * * *", ts.Spec.Schedule)
				testutil.AssertEqual(t, "suspend", false, ts.Spec.Suspend)
			},
		},
		{
			name:  "sets concurrency policy if provided",
			space: spaceName,
			args:  []string{appName, jobName, command, "--concurrency-policy", "Forbid", "--async"},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					Apps(spaceName).
					Create(ctx, app, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(
					t,
					"output",
					"Job my-job created.\nWaiting for Job to become ready asynchronously\n",
					buffer.String())
				client := fakeclient.Get(ctx)
				ts, err := client.KfV1alpha1().
					TaskSchedules(spaceName).
					Get(context.Background(), jobName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "concurrencyPolicy", "Forbid", ts.Spec.ConcurrencyPolicy)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gomock.NewController(t)

			cmd := taskschedules.NewCreateJobCommand(&config.KfParams{
				Space: tc.space,
			})

			var buffer bytes.Buffer

			ctx := fakeinjection.WithInjection(context.Background(), t)
			ctx = configlogging.SetupLogger(ctx, &buffer)

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
