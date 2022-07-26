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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fakeclient "github.com/google/kf/v2/pkg/client/kf/injection/client/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/taskschedules"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	krand "k8s.io/apimachinery/pkg/util/rand"
	ktesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func generateNameReactor(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
	obj := action.(ktesting.CreateAction).GetObject().(controllerutil.Object)
	if obj.GetName() == "" && obj.GetGenerateName() != "" {
		obj.SetName(fmt.Sprintf("%s%s", obj.GetGenerateName(), krand.String(8)))
	}
	return false, nil, nil
}

func TestRunJob(t *testing.T) {
	t.Parallel()

	const (
		spaceName = "my-space"
		appName   = "my-app"
		jobName   = "my-job"
		command   = "sleep 123"
		cpu       = "2G"
		disk      = "1Gi"
		memory    = "3Gi"
	)

	var (
		taskSchedule = &v1alpha1.TaskSchedule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: spaceName,
			},
			Spec: v1alpha1.TaskScheduleSpec{
				Schedule: "* * * * *",
				TaskTemplate: v1alpha1.TaskSpec{
					AppRef: v1.LocalObjectReference{
						Name: appName,
					},
					Command: command,
					Disk:    disk,
					CPU:     cpu,
					Memory:  memory,
				},
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
			expectErr: errors.New("accepts 1 arg(s), received 0"),
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
			name:  "creates Task",
			space: spaceName,
			args:  []string{jobName},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, taskSchedule, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				output := buffer.String()
				testutil.AssertRegexp(
					t,
					"stdout",
					"Task .* is submitted successfully for execution\\.",
					output)
				taskName := strings.Fields(output)[1]
				client := fakeclient.Get(ctx)
				task, err := client.KfV1alpha1().
					Tasks(spaceName).
					Get(ctx, taskName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "app", appName, task.Spec.AppRef.Name)
				testutil.AssertEqual(t, "command", command, task.Spec.Command)
			},
		},
		{
			name:  "overrides resource flags",
			space: spaceName,
			args: []string{
				jobName,
				"--cpu-cores", "5G",
				"--memory-limit", "6G",
				"--disk-quota", "7G",
			},
			setup: func(ctx context.Context, t *testing.T) {
				client := fakeclient.Get(ctx)
				client.KfV1alpha1().
					TaskSchedules(spaceName).
					Create(ctx, taskSchedule, metav1.CreateOptions{})
			},
			assert: func(ctx context.Context, t *testing.T, buffer *bytes.Buffer, err error) {
				output := buffer.String()
				testutil.AssertRegexp(
					t,
					"stdout",
					"Task .* is submitted successfully for execution\\.",
					output)
				taskName := strings.Fields(output)[1]
				client := fakeclient.Get(ctx)
				task, err := client.KfV1alpha1().
					Tasks(spaceName).
					Get(ctx, taskName, metav1.GetOptions{})
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "app", appName, task.Spec.AppRef.Name)
				testutil.AssertEqual(t, "command", command, task.Spec.Command)
				testutil.AssertEqual(t, "cpu", "5G", task.Spec.CPU)
				testutil.AssertEqual(t, "memory", "6Gi", task.Spec.Memory)
				testutil.AssertEqual(t, "disk", "7Gi", task.Spec.Disk)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gomock.NewController(t)

			cmd := taskschedules.NewRunJobCommand(&config.KfParams{
				Space: tc.space,
			})

			var buffer bytes.Buffer

			ctx := fakeinjection.WithInjection(context.Background(), t)

			// The fake client does not automatically generate names, so we add
			// a reactor to mimic that behavior.
			client := fakeclient.Get(ctx)
			client.PrependReactor("create", "tasks", generateNameReactor)

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
