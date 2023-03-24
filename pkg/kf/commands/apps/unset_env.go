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

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewUnsetEnvCommand creates a SetEnv command.
func NewUnsetEnvCommand(p *config.KfParams, client apps.Client) *cobra.Command {
	var async utils.AsyncIfStoppedFlags

	cmd := &cobra.Command{
		Use:   "unset-env APP_NAME ENV_VAR_NAME",
		Short: "Delete an environment variable on an App.",
		Long: `
		Removes an environment variable by name from an App.
		Any existing environment variable(s) on the App with the same name will be removed.

		Environment variables that are set by Kf or on the Space will be unaffected.

		Apps will be updated without downtime.
		`,
		Example:           `kf unset-env myapp FOO`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]
			name := args[1]

			app, err := client.Transform(cmd.Context(), p.Space, appName, func(app *v1alpha1.App) error {
				kfapp := (*apps.KfApp)(app)
				kfapp.DeleteEnvVars([]string{name})

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to unset environment variable on App: %s", err)
			}

			stopped := app != nil && app.Spec.Instances.Stopped
			action := fmt.Sprintf("Unsetting environment variable on App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(stopped, cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	return cmd
}
