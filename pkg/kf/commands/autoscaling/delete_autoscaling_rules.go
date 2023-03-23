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
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewDeleteAutoscalingRules command deletes all autoscaling rules for an App.
func NewDeleteAutoscalingRules(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncIfStoppedFlags

	cmd := &cobra.Command{
		Use:               "delete-autoscaling-rules APP_NAME",
		Short:             "Delete all autoscaling rules for App and disable autoscaling.",
		Example:           `kf delete-autoscaling-rules myapp`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Autoscaling.Rules = nil

				return nil
			}

			app, err := client.Transform(cmd.Context(), p.Space, appName, mutator)
			if err != nil {
				return fmt.Errorf("failed to delete all autoscaling rules for App: %s", err)
			}

			stopped := app != nil && (app.Spec.Instances.Stopped || !app.Spec.Instances.Autoscaling.Enabled)
			action := fmt.Sprintf("Deleting all autoscaling rules for App %q in Space %q", appName, p.Space)
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
