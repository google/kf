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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appsfake "github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/tasks"
	tasksfake "github.com/google/kf/v2/pkg/kf/tasks/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestRunTask(t *testing.T) {
	const (
		appName   = "my-app"
		spaceName = "my-space"
		taskName  = "my-task"
		command   = "my-command"
	)

	sampleApp := &v1alpha1.App{
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Args: []string{},
						},
					},
				},
			},
		},
		Status: v1alpha1.AppStatus{
			Tasks: v1alpha1.AppTaskStatus{
				UpdateRequests: 1,
			},
		},
	}
	sampleTask := &v1alpha1.Task{
		Spec: v1alpha1.TaskSpec{
			AppRef: corev1.LocalObjectReference{
				Name: appName,
			},
		},
	}

	type fakes struct {
		apps  *appsfake.FakeClient
		tasks *tasksfake.FakeClient
	}

	t.Parallel()

	for tn, tc := range map[string]struct {
		Space     string
		Args      []string
		Setup     func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient)
		expectErr error
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"missing app name": {
			expectErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"wrong number of args": {
			Args:      []string{appName, "example.com"},
			expectErr: errors.New("accepts 1 arg(s), received 2"),
		},
		"no target space": {
			Args:      []string{appName},
			expectErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"Get App fails": {
			Space: spaceName,
			Args:  []string{appName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					Get(gomock.Any(), spaceName, appName).
					Return(nil, errors.New("Unable to get App"))
			},
			expectErr: errors.New("failed to run-task on App: Unable to get App"),
		},
		"missing starter command in both App and Task": {
			Space: spaceName,
			Args:  []string{appName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					Get(gomock.Any(), spaceName, appName).
					Return(sampleApp, nil)
			},
			expectErr: errors.New("Start command not found on App or Task"),
		},
		"create Task fails": {
			Space: spaceName,
			Args:  []string{appName, "--command", command},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					Get(gomock.Any(), spaceName, appName).
					Return(sampleApp, nil)
				fakeTasks.EXPECT().
					Create(gomock.Any(), spaceName, gomock.Any()).
					Return(nil, errors.New("Create Task failed"))
			},
			expectErr: errors.New("Create Task failed"),
		},
		"create Task succeeds with custom Task name": {
			Space: spaceName,
			Args:  []string{appName, "--command", command, "--name", taskName},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					Get(gomock.Any(), spaceName, appName).
					Return(sampleApp, nil)
				fakeTasks.EXPECT().
					Create(gomock.Any(), spaceName, gomock.Any()).
					Return(sampleTask, nil)
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"create Task succeeds with auto generated Task name": {
			Space: spaceName,
			Args:  []string{appName, "--command", command},
			Setup: func(t *testing.T, fakeTasks *tasksfake.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					Get(gomock.Any(), spaceName, appName).
					Return(sampleApp, nil)
				fakeTasks.EXPECT().
					Create(gomock.Any(), spaceName, gomock.Any()).
					Return(sampleTask, nil)
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var buffer bytes.Buffer

			aClient := appsfake.NewFakeClient(ctrl)
			tClient := tasksfake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, tClient, aClient)
			}

			cmd := tasks.NewRunTaskCommand(
				&config.KfParams{
					Space: tc.Space,
				},
				tClient,
				aClient)

			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()

			if tc.expectErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, gotErr)
			}

			if tc.Assert != nil {
				tc.Assert(t, &buffer, gotErr)
			}

			if gotErr != nil {
				return
			}

		})
	}
}
