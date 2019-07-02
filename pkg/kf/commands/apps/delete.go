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
	"log"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates a delete command.
func NewDeleteCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:     "delete APP_NAME",
		Short:   "Delete an existing app",
		Example: `  kf delete myapp`,
		Args:    cobra.ExactArgs(1),
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			if err := appsClient.DeleteInForeground(p.Namespace, appName); err != nil {
				return err
			}

			log.Printf("app %q has been successfully deleted\n", appName)
			return nil
		},
	}

	return deleteCmd
}
