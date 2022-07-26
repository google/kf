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
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

func TestIntegration_TaskSchedules(t *testing.T) {
	appName := fmt.Sprintf("integration-taskschedule-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/echo"

	// This test needs more time because it pushes an App with a V2 buildpack.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	t.Cleanup(cancel)

	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		kf.CachePushV2(ctx, appName, filepath.Join(integration.RootDir(ctx, t), appPath), "--task")
		ctx = integration.ContextWithApp(ctx, appName)

		app, ok := kf.Apps(ctx)[appName]
		testutil.AssertEqual(t, "app presence", true, ok)
		testutil.AssertEqual(t, "app instances", "stopped", app.Instances)

		t.Run("scheduled tasks", func(t *testing.T) {
			t.Run("manual job execution", func(t *testing.T) {
				t.Parallel()

				jobName := "manual-job"
				// Create a Job.
				kf.RunCommand(ctx, "create-job", appName, jobName, "echo MANUAL JOB COMPLETE")

				// Assert the Job is listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) == nil {
					t.Fatalf("jobs failed to list created Job")
				}

				// Run the Job manually.
				kf.RunCommand(ctx, "run-job", jobName)
				tasks := findJobTasks(ctx, kf, appName, jobName)

				// Assert the Task is created and runs successfully.
				kf.VerifyTaskLogOutput(ctx, appName, "MANUAL JOB COMPLETE", 120*time.Second)
				if len(tasks) != 1 {
					t.Fatalf("run-job failed to create Task")
				}

				// Delete the Job.
				kf.RunCommand(ctx, "delete-job", jobName)

				// Assert the Job is no longer listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) != nil {
					t.Fatalf("delete-job failed to delete the Job")
				}
			})

			t.Run("scheduled job execution", func(t *testing.T) {
				t.Parallel()

				jobName := "scheduled-job-always"
				// Create a Job.
				kf.RunCommand(ctx, "create-job", appName, jobName, "echo SCHEDULED JOB COMPLETE")

				// Assert the Job is listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) == nil {
					t.Fatalf("jobs failed to list created Job")
				}
				// Schedule the Job for repeated execution.
				kf.RunCommand(ctx, "schedule-job", jobName, "* * * * *")

				// Assert the Job's schedule is listed by `kf job-schedules`.
				if findJobSchedule(ctx, kf, appName, jobName) == nil {
					t.Fatalf("job-schedules failed to list scheduled Job")
				}

				// Assert the scheduled Task is created and runs successfully.
				kf.VerifyTaskLogOutput(ctx, appName, "SCHEDULED JOB COMPLETE", 180*time.Second)
				tasks := findJobTasks(ctx, kf, appName, jobName)
				if len(tasks) == 0 {
					t.Fatalf("schedule-job failed to automatically create Tasks")
				}

				// Delete the Job's schedule.
				kf.RunCommand(ctx, "delete-job-schedule", jobName)

				// Assert the Job's schedule is _not_ listed by `kf job-schedules`.
				if findJobSchedule(ctx, kf, appName, jobName) != nil {
					t.Fatalf("job-schedules listed Job which is no longer scheduled")
				}

				// Assert the Job is no longer creating scheduled Tasks.
				time.Sleep(1 * time.Minute)
				nextTasks := findJobTasks(ctx, kf, appName, jobName)
				if len(nextTasks) != len(tasks) {
					t.Fatalf("delete-job-schedule failed to stop creating Tasks")
				}

				// Delete the Job.
				kf.RunCommand(ctx, "delete-job", jobName)

				// Assert the Job is no longer listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) != nil {
					t.Fatalf("delete-job failed to delete the Job")
				}
			})

			t.Run("concurrency policy replace", func(t *testing.T) {
				t.Parallel()

				jobName := "scheduled-job-replace"
				// Create a Job.
				kf.RunCommand(ctx, "create-job", appName, jobName, "sleep 1000", "--concurrency-policy", "Replace")

				// Assert the Job is listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) == nil {
					t.Fatalf("jobs failed to list created Job")
				}
				// Schedule the Job for repeated execution.
				kf.RunCommand(ctx, "schedule-job", jobName, "* * * * *")

				// Assert the scheduled Task is created and cancelled by the next schedule execution.
				time.Sleep(3 * time.Minute)
				tasks := findJobTasks(ctx, kf, appName, jobName)
				if len(tasks) < 2 {
					t.Fatalf("schedule-job failed to automatically create Tasks")
				}
				if tasks[0].Reason != "TaskRunCancelled" {
					t.Fatalf("expected first task to be cancelled, got: %v", tasks)
				}

				// Delete the Job.
				kf.RunCommand(ctx, "delete-job", jobName)

				// Assert the Job is no longer listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) != nil {
					t.Fatalf("delete-job failed to delete the Job")
				}
			})

			t.Run("concurrency policy forbid", func(t *testing.T) {
				t.Parallel()

				jobName := "scheduled-job-forbid"
				// Create a Job.
				kf.RunCommand(ctx, "create-job", appName, jobName, "sleep 1000", "--concurrency-policy", "Forbid")

				// Assert the Job is listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) == nil {
					t.Fatalf("jobs failed to list created Job")
				}
				// Schedule the Job for repeated execution.
				kf.RunCommand(ctx, "schedule-job", jobName, "* * * * *")

				// Assert the scheduled Task is created and further scheduled executions are skipped.
				time.Sleep(3 * time.Minute)
				tasks := findJobTasks(ctx, kf, appName, jobName)
				if len(tasks) != 1 {
					t.Fatalf("expected only 1 running task execution, got: %v", tasks)
				}

				// Delete the Job.
				kf.RunCommand(ctx, "delete-job", jobName)

				// Assert the Job is no longer listed by `kf jobs`.
				if findJob(ctx, kf, appName, jobName) != nil {
					t.Fatalf("delete-job failed to delete the Job")
				}
			})
		})
	})
}

func findJob(ctx context.Context, kf *integration.Kf, appName, jobName string) *integration.JobInfo {
	jobs := kf.Jobs(ctx, appName)
	for _, j := range jobs {
		if j.Name == jobName {
			return &j
		}
	}
	return nil
}

func findJobSchedule(ctx context.Context, kf *integration.Kf, appName, jobName string) *integration.JobInfo {
	jobSchedules := kf.JobSchedules(ctx, appName)
	for _, js := range jobSchedules {
		if js.Name == jobName {
			return &js
		}
	}
	return nil
}

func findJobTasks(ctx context.Context, kf *integration.Kf, appName, jobName string) []integration.TaskInfo {
	tasks := kf.Tasks(ctx, appName)
	var results []integration.TaskInfo
	for _, t := range tasks {
		if strings.HasPrefix(t.Name, jobName) {
			results = append(results, t)
		}
	}
	return results
}
