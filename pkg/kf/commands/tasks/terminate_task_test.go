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

package tasks_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kffake "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/commands/tasks"
	tasksfake "github.com/google/kf/v2/pkg/kf/tasks/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	duckv1beta1 "knative.dev/pkg/apis"
)

func TestTerminateTask(t *testing.T) {
	const (
		appName   = "my-app"
		spaceName = "my-space"
		taskID    = "1"
		command   = "my-command"
		taskName  = "my-app-abc"
	)

	sampleTask := &v1alpha1.Task{}

	sampleTask.Name = taskName
	sampleTask.Namespace = spaceName
	sampleTask.APIVersion = "kf.dev/v1alpha1"
	sampleTask.Kind = "Task"
	sampleTask.Labels = map[string]string{
		v1alpha1.NameLabel:    appName,
		v1alpha1.VersionLabel: taskID,
	}
	sampleTask.Spec.AppRef = corev1.LocalObjectReference{
		Name: appName,
	}

	t.Parallel()
	for tn, tc := range map[string]struct {
		Space           string
		Args            []string
		Setup           func(t *testing.T, fakeTasks *tasksfake.FakeClient)
		expectErr       error
		expectedStrings []string
		Assert          func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"missing App name or Task name": {
			expectErr: errors.New("accepts between 1 and 2 arg(s), received 0"),
		},
		"wrong number of args": {
			Args:      []string{appName, taskID, "example.com"},
			expectErr: errors.New("accepts between 1 and 2 arg(s), received 3"),
		},
		"no target Space": {
			Args:      []string{appName, taskID},
			expectErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"getting Task fails": {
			Space: spaceName,
			Args:  []string{taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				fakeTasks.EXPECT().
					Get(gomock.Any(), spaceName, taskName).
					Return(nil, errors.New("Unable to get Task"))
			},
			expectErr: errors.New("Unable to get Task"),
		},
		"can't terminate completed Task": {
			Space: spaceName,
			Args:  []string{taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				completedTask := sampleTask.DeepCopy()
				completedTask.Status.Conditions = append(completedTask.Status.Conditions, duckv1beta1.Condition{
					Type:   duckv1beta1.ConditionSucceeded,
					Status: corev1.ConditionTrue,
				})

				fakeTasks.EXPECT().
					Get(gomock.Any(), spaceName, taskName).
					Return(completedTask, nil)
			},
			expectedStrings: []string{"Can't terminate completed Task"},
		},
		"can't terminate terminated Task": {
			Space: spaceName,
			Args:  []string{taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				terminatedTask := sampleTask.DeepCopy()
				terminatedTask.Spec.Terminated = true

				fakeTasks.EXPECT().
					Get(gomock.Any(), spaceName, taskName).
					Return(terminatedTask, nil)
			},
			expectedStrings: []string{"Can't terminate terminated Task"},
		},
		"terminate Task fails": {
			Space: spaceName,
			Args:  []string{taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				fakeTasks.EXPECT().
					Get(gomock.Any(), spaceName, taskName).
					Return(sampleTask, nil)
				fakeTasks.EXPECT().
					Transform(gomock.Any(), spaceName, taskName, gomock.Any()).
					Return(nil, errors.New("Unable to update Task"))
			},
			expectErr: errors.New("Failed to terminate Task: Unable to update Task"),
		},
		"terminate Task by Task name succeeds": {
			Space: spaceName,
			Args:  []string{taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				fakeTasks.EXPECT().
					Get(gomock.Any(), spaceName, taskName).
					Return(sampleTask, nil)
				fakeTasks.EXPECT().
					Transform(gomock.Any(), spaceName, sampleTask.Name, gomock.Any()).
					Return(nil, nil)
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"terminate Task by App name and Task ID succeeds": {
			Space: spaceName,
			Args:  []string{appName, taskID},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient) {
				fakeTasks.EXPECT().
					Transform(gomock.Any(), spaceName, taskName, gomock.Any()).
					Return(nil, nil)
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var buffer bytes.Buffer

			kfClientSet := kffake.NewSimpleClientset(sampleTask)
			kfClient := kfClientSet.KfV1alpha1()
			tClient := tasksfake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, tClient)
			}

			ctx := configlogging.SetupLogger(context.Background(), &buffer)

			cmd := tasks.NewTerminateTaskCommand(
				&config.KfParams{
					Space: tc.Space,
				},
				tClient,
				kfClient)

			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)
			cmd.SetContext(ctx)

			gotErr := cmd.Execute()

			if tc.expectErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, gotErr)
			}

			testutil.AssertContainsAll(t, buffer.String(), tc.expectedStrings)

			if tc.Assert != nil {
				tc.Assert(t, &buffer, gotErr)
			}

			if gotErr != nil {
				return
			}

		})
	}
}
