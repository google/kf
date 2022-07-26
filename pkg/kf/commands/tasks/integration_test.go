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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

func TestIntegration_Tasks(t *testing.T) {
	appName := fmt.Sprintf("integration-task-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/echo"
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		integration.WithTaskApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
			app, ok := kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			testutil.AssertEqual(t, "app instances", "stopped", app.Instances)

			// Run task.
			kf.RunCommand(ctx, "run-task", appName, "--command", "\"sleep 120\"")
			task := findTask(ctx, t, kf, appName)

			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			for task.Reason != "Running" && ctx.Err() == nil {
				task = findTask(ctx, t, kf, appName)
				integration.Logf(t, "Waiting for Task %s to be running", task.Name)
			}

			if err := ctx.Err(); err != nil {
				t.Fatalf("Context error: %v", err)
			}

			// Terminate task.
			kf.RunCommand(ctx, "terminate-task", task.Name)
			terminatedTask := findTask(ctx, t, kf, appName)

			ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			for terminatedTask.Status != "False" && ctx.Err() == nil {
				terminatedTask = findTask(ctx, t, kf, appName)
				integration.Logf(t, "Waiting for Task %s to be terminated", terminatedTask.Name)
			}

			if err := ctx.Err(); err != nil {
				t.Fatalf("Context error: %v", err)
			}

			testutil.AssertEqual(t, "task reason", "TaskRunCancelled", terminatedTask.Reason)
		})
	})
}

func findTask(ctx context.Context, t *testing.T, kf *integration.Kf, appName string) (task integration.TaskInfo) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// Check `kf tasks APP_NAME` for task
	tasks := kf.Tasks(ctx, appName)

	if len(tasks) == 0 {
		t.Fatalf("No task is running")
	}
	if len(tasks) > 1 {
		t.Fatalf("More than one task is running")
	}

	return tasks[0]
}
