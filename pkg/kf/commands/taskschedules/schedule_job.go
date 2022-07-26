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

	"github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewScheduleJobCommand schedules the specified TaskSchedule with the given
// cron expression.
func NewScheduleJobCommand(p *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "schedule-job JOB_NAME SCHEDULE",
		Short:        "Schedule the Job for execution on a cron schedule.",
		Example:      `kf schedule-job my-job "* * * * *"`,
		Args:         cobra.ExactArgs(2),
		Long:         `The schedule-job sub-command lets operators schedule a Job for execution on a cron schedule.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			jobName := args[0]
			schedule := args[1]

			if _, err := cron.ParseStandard(schedule); err != nil {
				return fmt.Errorf("Schedule %q is not a valid cron schedule: %s", schedule, err)
			}

			client := client.Get(cmd.Context())
			ts, err := client.KfV1alpha1().
				TaskSchedules(p.Space).
				Get(cmd.Context(), jobName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !ts.Spec.Suspend {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"%s: Scheduling an already scheduled Job. This will overwrite the existing schedule.",
					warningText)
			}

			ts.Spec.Schedule = schedule
			ts.Spec.Suspend = false

			_, err = client.KfV1alpha1().
				TaskSchedules(p.Space).
				Update(cmd.Context(), ts, metav1.UpdateOptions{})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Job %s scheduled.\n", ts.Name)
			return nil
		},
	}
	return cmd
}
