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
	"strconv"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	"knative.dev/pkg/ptr"
)

// NewCreateAutoscalingRule command enables autoscaling for an App.
func NewCreateAutoscalingRule(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncIfStoppedFlags

	cmd := &cobra.Command{
		Use:   "create-autoscaling-rule APP RULE_TYPE MIN_THRESHOLD MAX_THRESHOLD",
		Short: "Create autoscaling rule for App.",
		Long: `
		Create an autoscaling rule for App.

		The only supported rule type is CPU. It is the target
		percentage. It is calculated by taking the average of MIN_THRESHOLD
		and MAX_THRESHOLD.

		The range of MIN_THRESHOLD and MAX_THRESHOLD is 1 to 100 (percent).
		`,
		Example: `
		# Scale myapp based on CPU load targeting 50% utilization (halfway between 20 and 80)
		kf create-autoscaling-rule myapp CPU 20 80
		`,
		Args:              cobra.ExactArgs(4),
		ValidArgsFunction: completion.AppCompletionFn(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName, ruleType := args[0], v1alpha1.GetAutoscalingRuleType(args[1])

			minThreshold, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("min threshold has to be an integer: %s", err)
			}

			maxThreshold, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("max threshold has to be an integer: %s", err)
			}

			// Validation on rules are done on the server side.
			// For v1, there can be only one rule and rule type has to be CPU.
			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Autoscaling.Rules =
					append(app.Spec.Instances.Autoscaling.Rules,
						v1alpha1.AppAutoscalingRule{
							RuleType: ruleType,
							Target:   ptr.Int32(int32((minThreshold + maxThreshold) / 2)),
						})
				return nil
			}

			app, err := client.Transform(cmd.Context(), p.Space, appName, mutator)
			if err != nil {
				return fmt.Errorf("failed to add autoscaling rule for App: %s", err)
			}

			stopped := app != nil && (app.Spec.Instances.Stopped || !app.Spec.Instances.Autoscaling.Enabled)
			action := fmt.Sprintf("Creating autoscaling rule for App %q in Space %q", appName, p.Space)
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
