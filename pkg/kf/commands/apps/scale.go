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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/spf13/cobra"
)

// NewScaleCommand creates a command capable of scaling an app.
func NewScaleCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var (
		instances    int
		autoscaleMin int
		autoscaleMax int
	)

	cmd := &cobra.Command{
		Use:   "scale APP_NAME",
		Short: "Change or view the instance count for an app",
		Example: `
		# Display current scale settings
		kf scale myapp
		# Scale to exactly 3 instances
		kf scale myapp --instances 3
		# Scale to at least 3 instances
		kf scale myapp --min 3
		# Scale between 0 and 5 instances
		kf scale myapp --max 5
		# Scale between 3 and 5 instances depending on traffic
		kf scale myapp --min 3 --max 5
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			cmd.SilenceUsage = true

			appName := args[0]

			if instances < 0 && autoscaleMin < 0 && autoscaleMax < 0 {
				// Display current scaling properties.
				app, err := client.Get(p.Namespace, appName)
				if err != nil {
					return fmt.Errorf("failed to get app: %s", err)
				}
				describe.AppSpecInstances(cmd.OutOrStderr(), app.Spec.Instances)

				return nil
			}

			// Manipulate the scaling

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Min = nil
				app.Spec.Instances.Max = nil
				app.Spec.Instances.Exactly = nil

				if instances >= 0 {
					// Exact
					app.Spec.Instances.Exactly = &instances
				}

				if autoscaleMin >= 0 {
					// Min is set
					app.Spec.Instances.Min = &autoscaleMin
				}

				if autoscaleMax >= 0 {
					// Max is set
					app.Spec.Instances.Max = &autoscaleMax
				}

				if err := app.Spec.Instances.Validate(context.Background()); err != nil {
					return err
				}

				describe.AppSpecInstances(cmd.OutOrStderr(), app.Spec.Instances)

				return nil
			}

			if err := client.Transform(p.Namespace, appName, mutator); err != nil {
				return fmt.Errorf("failed to scale app: %s", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Scaling app %q %s", appName, utils.AsyncLogSuffix)
			return nil
		},
	}

	cmd.Flags().IntVarP(
		&instances,
		"instances",
		"i",
		-1,
		"Number of instances.",
	)

	cmd.Flags().IntVar(
		&autoscaleMin,
		"min",
		-1,
		"Minimum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to 0.",
	)

	cmd.Flags().IntVar(
		&autoscaleMax,
		"max",
		-1,
		"Maximum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to ∞.",
	)

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
