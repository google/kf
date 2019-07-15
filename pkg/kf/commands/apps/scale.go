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
	"errors"
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/kfapps"
	"github.com/spf13/cobra"
)

// NewScaleCommand creates a command capable of scaling an app.
func NewScaleCommand(
	p *config.KfParams,
	client kfapps.Client,
) *cobra.Command {
	var (
		instances    int
		autoscaleMin int
		autoscaleMax int
	)

	var scale = &cobra.Command{
		Use:   "scale APP_NAME",
		Short: "Change the instance count for an app.",
		Example: `
  kf scale myapp --i 3 # Scale to exactly 3 instances
  kf scale myapp --instances 3 # Scale to exactly 3 instances
  kf scale myapp --min 3 # Autoscaler won't scale below 3 instances
  kf scale myapp --max 5 # Autoscaler won't scale above 5 instances
  kf scale myapp --min 3 --max 5 # Autoscaler won't below 3 or above 5 instances
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			if instances < 0 && autoscaleMin < 0 && autoscaleMax < 0 {
				return errors.New("--instances, --min, or --max flag are required")
			}

			appName := args[0]

			cmd.SilenceUsage = true

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

				fmt.Fprintf(cmd.OutOrStderr(), app.Spec.Instances.PrettyPrint())

				return nil
			}

			if err := client.Transform(p.Namespace, appName, mutator); err != nil {
				return fmt.Errorf("failed to scale app: %s", err)
			}

			return nil
		},
	}

	scale.Flags().IntVarP(
		&instances,
		"instances",
		"i",
		-1,
		"Number of instances.",
	)

	scale.Flags().IntVar(
		&autoscaleMin,
		"min",
		-1,
		"Minimum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to 0.",
	)

	scale.Flags().IntVar(
		&autoscaleMax,
		"max",
		-1,
		"Maximum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to ∞.",
	)

	return scale
}
