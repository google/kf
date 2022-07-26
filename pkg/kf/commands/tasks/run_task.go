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
	"errors"
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/tasks"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
)

// NewRunTaskCommand creates a short-running Task run on a given App.
func NewRunTaskCommand(p *config.KfParams, client tasks.Client, appClient apps.Client) *cobra.Command {
	var (
		command       string
		name          string
		resourceFlags utils.ResourceFlags
	)
	cmd := &cobra.Command{
		Use:               "run-task APP_NAME",
		Short:             "Run a short-lived Task on the App.",
		Example:           `kf run-task my-app --command "sleep 100" --name my-task`,
		Args:              cobra.ExactArgs(1),
		Long:              `The run-task sub-command lets operators run a short-lived Task on the App.`,
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]
			app, appErr := appClient.Get(ctx, p.Space, appName)

			if appErr != nil {
				return fmt.Errorf("failed to run-task on App: %s", appErr)
			}

			if len(app.Spec.Template.Spec.Containers[0].Args) == 0 && len(command) == 0 {
				return errors.New("Start command not found on App or Task")
			}

			desiredTask := &v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: appName + "-",
					Namespace:    p.Space,
					OwnerReferences: []metav1.OwnerReference{
						*kmeta.NewControllerRef(app),
					},
				},
				Spec: v1alpha1.TaskSpec{
					AppRef: corev1.LocalObjectReference{
						Name: appName,
					},
					// CPU is not converted to SI because it's not a normal CF field
					// and is therefore expected to be in SI to begin with.
					CPU:     resourceFlags.CPU(),
					Memory:  manifest.CFToSIUnits(resourceFlags.Memory()),
					Disk:    manifest.CFToSIUnits(resourceFlags.Disk()),
					Command: command,
				},
			}

			if len(name) > 0 {
				desiredTask.Spec.DisplayName = name
			}

			task, err := client.Create(ctx, p.Space, desiredTask)

			if err != nil {
				return err
			}

			logging.FromContext(ctx).Infof("Task %s is submitted successfully for execution.", task.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(
		&command,
		"command",
		"c",
		"",
		"Command to execute on the Task.",
	)

	cmd.Flags().StringVar(
		&name,
		"name",
		"",
		"Display name to give the Task (auto generated if omitted).",
	)

	resourceFlags.Add(cmd)

	return cmd
}
