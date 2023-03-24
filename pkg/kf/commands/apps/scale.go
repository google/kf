// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apps

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

// NewScaleCommand creates a command capable of scaling an app.
func NewScaleCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var (
		async utils.AsyncIfStoppedFlags

		instances int32
	)

	cmd := &cobra.Command{
		Use:   "scale APP_NAME",
		Short: "Change the horizontal or vertical scale of an App without downtime.",
		Long: `
		Scaling an App will change the number of desired instances and/or the
		requested resources for each instance.

		Instances are replaced one at a time, always ensuring that the desired
		number of instances are healthy. This property is upheld by running one
		additional instance of the App and swapping it out for an old instance.

		The operation completes once all instances have been replaced.
		`,
		Example: `
		# Display current scale settings
		kf scale myapp
		# Scale to exactly 3 instances
		kf scale myapp --instances 3
		`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			if instances <= 0 {
				// Display current scaling properties.
				app, err := client.Get(cmd.Context(), p.Space, appName)
				if err != nil {
					return fmt.Errorf("failed to get App: %s", err)
				}
				describe.AppSpecInstances(cmd.OutOrStderr(), app.Spec.Instances)
				return nil
			}

			// Manipulate the scaling
			mutator := func(app *v1alpha1.App) error {
				if app.Spec.Instances.Autoscaling.RequiresHPA() {
					utils.SuggestNextAction(utils.NextAction{
						Description: "Disable autoscaling",
						Commands: []string{
							fmt.Sprintf("kf disable-autoscaling %s", appName),
						},
					})
					return fmt.Errorf("cannot scale App manually when autoscaling is turned on")
				}

				app.Spec.Instances.Replicas = nil

				if instances > 0 {
					// Exact
					app.Spec.Instances.Replicas = &instances
				}

				if err := app.Spec.Instances.Validate(context.Background()); err != nil {
					return err
				}

				describe.AppSpecInstances(cmd.OutOrStderr(), app.Spec.Instances)

				return nil
			}

			app, err := client.Transform(cmd.Context(), p.Space, appName, mutator)
			if err != nil {
				return fmt.Errorf("failed to scale App: %s", err)
			}

			stopped := app != nil && app.Spec.Instances.Stopped
			action := fmt.Sprintf("Scaling App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(stopped, cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	cmd.Flags().Int32VarP(
		&instances,
		"instances",
		"i",
		-1,
		"Number of instances, must be >= 1.",
	)

	return cmd
}
