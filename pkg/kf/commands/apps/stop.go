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
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewStopCommand creates a command capable of stopping an app.
func NewStopCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:   "stop APP_NAME",
		Short: "Remove instances of a running App and stop network traffic.",
		Long: `
		Stopping an App will remove the associated Kubernetes Deployment and Service.
		This will also delete any historical revisions associated with the App.
		HTTP traffic being routed to the App will instead be routed to other Apps
		bound to the Route or will error with an 404 status code if no other Apps
		are bound.
		`,
		Example:           `kf stop myapp`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Stopped = true
				return nil
			}

			if _, err := client.Transform(cmd.Context(), p.Space, appName, mutator); err != nil {
				return fmt.Errorf("failed to stop App: %s", err)
			}

			action := fmt.Sprintf("Stopping App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	return cmd
}
