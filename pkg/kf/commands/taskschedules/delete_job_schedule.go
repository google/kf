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

	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	warningColor = color.New(color.FgHiYellow, color.Bold)
	warningText  = warningColor.Sprintf("WARNING")
)

// NewDeleteJobScheduleCommand suspends the given TaskSchedule.
func NewDeleteJobScheduleCommand(p *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete-job-schedule JOB_NAME",
		Short:   "Delete the schedule for a Job.",
		Example: `kf delete-job-schedule my-job`,
		// We accept an additional arg for compatibility with PCF cli; however,
		// we print a warning message alerting the user it is unused.
		Args: cobra.RangeArgs(1, 2),
		Long: `
		The delete-job-schedule sub-command lets operators suspend the Job by
		deleting the schedule.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			jobName := args[0]

			if len(args) > 1 {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"%s: Ignoring SCHEDULE-GUID arg. Kf only supports a single schedule per Job.",
					warningText)
			}

			client := client.Get(cmd.Context())
			ts, err := client.KfV1alpha1().TaskSchedules(p.Space).Get(cmd.Context(), jobName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if ts.Spec.Suspend {
				return fmt.Errorf("Job is already suspended.")
			}

			ts.Spec.Suspend = true

			_, err = client.KfV1alpha1().
				TaskSchedules(p.Space).
				Update(cmd.Context(), ts, metav1.UpdateOptions{})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Job schedule for %s deleted.\n", ts.Name)
			return nil
		},
	}
	return cmd
}
