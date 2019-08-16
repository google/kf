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
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/spf13/cobra"
)

// NewEnvCommand creates a Env command.
func NewEnvCommand(p *config.KfParams, appClient apps.Client) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:     "env APP_NAME",
		Short:   "List the names and values of the environment variables for an app",
		Example: `  kf env myapp`,
		Args:    cobra.ExactArgs(1),
		Long: `The env command gets the names and values of developer managed
		environment variables for an application.

		This command does not include environment variables that are set by kf
		such as VCAP_SERVICES or set by operators for all apps on the space.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]
			cmd.SilenceUsage = true

			app, err := appClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			kfapp := (*apps.KfApp)(app)
			describe.EnvVars(cmd.OutOrStdout(), kfapp.GetEnvVars())

			return nil
		},
	}

	return envCmd
}
