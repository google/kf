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
	"fmt"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewRestageCommand creates a command capable of restaging an app.
func NewRestageCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async bool

	cmd := &cobra.Command{
		Use:     "restage APP_NAME",
		Short:   "Rebuild and deploy using the last uploaded source code and current buildpacks",
		Example: `kf restage myapp`,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"rg"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			cmd.SilenceUsage = true

			app, err := client.Restage(p.Namespace, appName)
			if err != nil {
				return fmt.Errorf("failed to restage app: %s", err)
			}

			if !async {
				if err := client.DeployLogsForApp(cmd.OutOrStdout(), app); err != nil {
					return fmt.Errorf("failed to restage app: %s", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%q successfully restaged\n", appName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(
		&async,
		"async",
		"",
		false,
		"Don't wait for the restage to finish before returning.",
	)

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
