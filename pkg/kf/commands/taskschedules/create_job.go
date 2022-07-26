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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
)

// Placeholder cron expression to set schedule when user omits schedule flag.
// Suspend will be set to true alongside this schedule so it should never
// execute. Additionally, this schedule is the 30th of February which never
// occurs.
const placeholderCron = "0 0 30 2 *"

// NewCreateJobCommand creates a suspended TaskSchedule on a given App.
func NewCreateJobCommand(p *config.KfParams) *cobra.Command {
	var (
		resourceFlags     utils.ResourceFlags
		schedule          string
		concurrencyPolicy string
		async             utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:          "create-job APP_NAME JOB_NAME COMMAND",
		Short:        "Create a Job on the App.",
		Example:      `kf create-job my-app my-job "sleep 100"`,
		Args:         cobra.ExactArgs(3),
		Long:         `The create-job sub-command lets operators create a Job that can be run on a schedule or ad hoc.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]
			jobName := args[1]
			command := args[2]

			client := client.Get(ctx)

			app, err := client.KfV1alpha1().
				Apps(p.Space).
				Get(ctx, appName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get App: %s", err)
			}

			desiredTaskSchedule := &v1alpha1.TaskSchedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobName,
					Namespace: p.Space,
					OwnerReferences: []metav1.OwnerReference{
						*kmeta.NewControllerRef(app),
					},
				},
				Spec: v1alpha1.TaskScheduleSpec{
					Schedule:          placeholderCron,
					Suspend:           true,
					ConcurrencyPolicy: concurrencyPolicy,
					TaskTemplate: v1alpha1.TaskSpec{
						AppRef: corev1.LocalObjectReference{
							Name: appName,
						},
						CPU:     resourceFlags.CPU(),
						Memory:  manifest.CFToSIUnits(resourceFlags.Memory()),
						Disk:    manifest.CFToSIUnits(resourceFlags.Disk()),
						Command: command,
					},
				},
			}

			if schedule != "" {
				desiredTaskSchedule.Spec.Schedule = schedule
				desiredTaskSchedule.Spec.Suspend = false
			}

			taskSchedule, err := client.KfV1alpha1().
				TaskSchedules(p.Space).
				Create(ctx, desiredTaskSchedule, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create Job: %s", err)
			}

			logging.FromContext(ctx).Infof("Job %s created.", taskSchedule.Name)

			return async.WaitFor(
				ctx,
				cmd.OutOrStderr(),
				"Waiting for Job to become ready",
				time.Second,
				func() (bool, error) {
					ts, err := client.KfV1alpha1().
						TaskSchedules(p.Space).
						Get(ctx, jobName, metav1.GetOptions{})

					if err != nil {
						return false, err
					}
					return ts.Status.IsReady(), nil
				},
			)
		},
	}

	resourceFlags.Add(cmd)
	async.Add(cmd)

	// The default is left as "" here to determine if the schedule flag was
	// provided. When not provided the schedule is defaulted to placeholderCron.
	cmd.Flags().StringVarP(
		&schedule,
		"schedule",
		"s",
		"",
		"Cron schedule on which to execute the Job.",
	)

	cmd.Flags().StringVarP(
		&concurrencyPolicy,
		"concurrency-policy",
		"c",
		"Always",
		"Specifies how to treat concurrent executions of a Job: Always (default), Replace, or Forbid.",
	)

	return cmd
}
