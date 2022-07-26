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

package tasks

import (
	"fmt"

	kf "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"knative.dev/pkg/logging"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/tasks"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewTerminateTaskCommand terminates a running Task on a given App.
func NewTerminateTaskCommand(p *config.KfParams, client tasks.Client, kfClient kf.KfV1alpha1Interface) *cobra.Command {
	var (
		async utils.AsyncFlags
		retry utils.RetryFlags
	)
	cmd := &cobra.Command{
		Use:   "terminate-task {TASK_NAME | APP_NAME TASK_ID}",
		Short: "Terminate a running Task.",
		Example: `
# Terminate Task by Task name
kf terminate-task my-task-name

# Terminate Task by App name and Task ID
kf terminate-task my-app 1
`,
		Args:         cobra.RangeArgs(1, 2),
		Long:         `Allows operators to terminate a running Task on an App.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			var taskObjectName string
			var task *v1alpha1.Task
			if len(args) == 1 {
				taskObjectName = args[0]
				var taskErr error
				task, taskErr = client.Get(ctx, p.Space, taskObjectName)

				if taskErr != nil {
					return taskErr
				}
			} else {
				appName := args[0]
				taskID := args[1]
				selectors := map[string]string{
					v1alpha1.NameLabel:    appName,
					v1alpha1.VersionLabel: taskID,
				}
				listOptions := metav1.ListOptions{
					LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(selectors)),
				}

				taskList, err := kfClient.Tasks(p.Space).List(ctx, listOptions)

				if err != nil {
					return err
				}

				if len(taskList.Items) == 0 {
					return fmt.Errorf("No Task found for App %q and Task ID %q", appName, taskID)
				}

				task = &taskList.Items[0]
			}

			if v1alpha1.IsStatusFinal(task.Status.Status) {
				logger.Info("Can't terminate completed Task")
				return nil
			}

			if task.Spec.Terminated == true {
				logger.Info("Can't terminate terminated Task")
				return nil
			}

			mutator := func(task *v1alpha1.Task) error {
				task.Spec.Terminated = true
				return nil
			}

			if _, err := client.Transform(ctx, p.Space, task.Name, mutator); err != nil {
				return fmt.Errorf("Failed to terminate Task: %s", err)
			}

			logger.Infof("Task %q is successfully submitted for termination", task.Name)
			return nil
		},
	}
	async.Add(cmd)
	retry.AddRetryForK8sPropagation(cmd)

	return cmd
}
