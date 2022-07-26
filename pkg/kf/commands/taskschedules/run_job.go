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

package taskschedules

import (
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/reconciler/taskschedule/resources"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRunJobCommand runs the Task specified by the given TaskSchedule.
func NewRunJobCommand(p *config.KfParams) *cobra.Command {
	var (
		resourceFlags utils.ResourceFlags
	)

	cmd := &cobra.Command{
		Use:          "run-job JOB_NAME",
		Short:        "Run the Job once.",
		Example:      `kf run-job my-job`,
		Args:         cobra.ExactArgs(1),
		Long:         `The run-job sub-command lets operators run a Job once.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			jobName := args[0]

			client := client.Get(cmd.Context())
			ts, err := client.KfV1alpha1().
				TaskSchedules(p.Space).
				Get(cmd.Context(), jobName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			desiredTask := resources.MakeTask(ts, time.Now())
			// Generate random names for the Task to avoid conflicts between
			// manual executions and scheduled executions.
			desiredTask.Name = ""
			desiredTask.GenerateName = fmt.Sprintf("%s-", ts.Name)

			// Override TaskSchedule's resource flags if specified.
			if resourceFlags.CPU() != "" {
				desiredTask.Spec.CPU = resourceFlags.CPU()
			}
			if resourceFlags.Memory() != "" {
				desiredTask.Spec.Memory = manifest.CFToSIUnits(resourceFlags.Memory())
			}
			if resourceFlags.Disk() != "" {
				desiredTask.Spec.Disk = manifest.CFToSIUnits(resourceFlags.Disk())
			}

			task, err := client.KfV1alpha1().
				Tasks(p.Space).
				Create(cmd.Context(), desiredTask, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Task %s is submitted successfully for execution.\n",
				task.Name)
			return nil
		},
	}

	resourceFlags.Add(cmd)

	return cmd
}
