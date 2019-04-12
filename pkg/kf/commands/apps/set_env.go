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
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewSetEnvCommand creates a SetEnv command.
func NewSetEnvCommand(p *config.KfParams, c EnvironmentClient) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:   "set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE",
		Short: "Set an environment variable for an app",
		Args:  cobra.ExactArgs(3),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			name := args[1]
			value := args[2]

			cmd.SilenceUsage = true

			err := c.Set(
				appName,
				map[string]string{name: value},
				kf.WithSetEnvNamespace(p.Namespace),
			)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return envCmd
}
