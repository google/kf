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

package autoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewEnableAutoscaling command enables autoscaling for an App.
func NewEnableAutoscaling(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncIfStoppedFlags

	cmd := &cobra.Command{
		Use:   "enable-autoscaling APP_NAME",
		Short: "Enable autoscaling for App.",
		Long: `
		Enabling autoscaling creates a HorizontalPodAutoscaler (HPA) for an App. HPA
		controls the number of replicas the App should have based on CPU utilization.

		Autoscaler will only take effect after autoscaling limits are set and autoscaling
		rules are added.

		Set autoscaling limits by running:
			kf update-autoscaling-limits APP_NAME MIN_INSTANCES MAX_INSTANCES

		Add rules by running:
			kf create-autoscaling-rule APP_NAME RULE_TYPE MIN_THRESHOLD MAX_THRESHOLD
		`,
		Example:           `kf enable-autoscaling myapp`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			app, err := client.Get(cmd.Context(), p.Space, appName)
			if err != nil {
				return fmt.Errorf("failed to get App: %s", err)
			}

			// Autoscaling already enabled, display current scaling properties
			if app.Spec.Instances.Autoscaling.Enabled {
				describe.AppSpecAutoscaling(cmd.OutOrStderr(), &app.Spec.Instances.Autoscaling)
				return nil
			}

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Autoscaling.Enabled = true

				// Suggest follow up action after autoscalor is enabled.
				if app.Spec.Instances.Autoscaling.MaxReplicas == nil {
					utils.SuggestNextAction(utils.NextAction{
						Description: "Update autoscaling limits",
						Commands: []string{
							fmt.Sprintf("kf update-autoscaling-limits %s MIN-INSTANCE-LIMIT MAX-INSTANCE-LIMIT --space %s", app.Name, app.Namespace),
						},
					})
				}

				if len(app.Spec.Instances.Autoscaling.Rules) == 0 {
					utils.SuggestNextAction(utils.NextAction{
						Description: "Create autoscaling rule",
						Commands: []string{
							fmt.Sprintf("kf create-autoscaling-rule %s RULETYPE MIN-THRESHOLD MAX-THRESHOLD --space %s", app.Name, app.Namespace),
						},
					})
				}

				return nil
			}

			if _, err := client.Transform(cmd.Context(), p.Space, appName, mutator); err != nil {
				return fmt.Errorf("failed to enable autoscaling for App: %s", err)
			}

			stopped := app != nil && app.Spec.Instances.Stopped
			action := fmt.Sprintf("Enabling autoscaling for App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(stopped, cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
		SilenceUsage: true,
	}

	async.Add(cmd)

	return cmd
}
