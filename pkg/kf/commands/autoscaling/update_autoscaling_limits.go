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

// NewUpdateAutoscalingLimits command enables autoscaling for an App.
func NewUpdateAutoscalingLimits(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:   "update-autoscaling-limits APP_NAME MIN_INSTANCE_LIMIT MAX_INSTANCE_LIMIT",
		Short: "Update autoscaling limits for App.",
		Long:  "",
		Example: `
	# Set min instances to 1, max instances to 3 for myapp
	kf update-autoscaling-limits myapp 1 3
	`,
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.AppCompletionFn(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			minReplicas, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("min instances has to be an integer: %s", err)
			}

			maxReplicas, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("max instances has to be an integer: %s", err)
			}

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Autoscaling.MinReplicas = ptr.Int32(int32(minReplicas))
				app.Spec.Instances.Autoscaling.MaxReplicas = ptr.Int32(int32(maxReplicas))
				return nil
			}

			if _, err := client.Transform(cmd.Context(), p.Space, appName, mutator); err != nil {
				return fmt.Errorf("failed to update autoscaling limits for App: %s", err)
			}

			action := fmt.Sprintf("updating autoscaling limits for App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
		SilenceUsage: true,
	}

	async.Add(cmd)

	return cmd
}
