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

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/spf13/cobra"
)

// NewRestartCommand creates a command capable of restarting an app.
func NewRestartCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var restart = &cobra.Command{
		Use:   "restart APP_NAME",
		Short: "Restart stops the current pods and create new ones",
		Example: `
  kf restart myapp
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			cmd.SilenceUsage = true

			if err := client.Restart(p.Namespace, appName); err != nil {
				return fmt.Errorf("failed to restart app: %s", err)
			}

			return nil
		},
	}
	return restart
}
