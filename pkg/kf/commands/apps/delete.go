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
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates a delete command.
func NewDeleteCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete APP_NAME",
		Short:   "Delete an existing app",
		Example: `kf delete myapp`,
		Args:    cobra.ExactArgs(1),
		Long: `
		This command deletes an application from kf.

		Things that won't be deleted:

		* source code
		* application images
		* routes
		* service instances

		Things that will be deleted:

		* builds
		* bindings

		The delete occurs asynchronously. Apps are often deleted shortly after the
		delete command is called, but may live on for a while if:

		* there are still connections waiting to be served
		* bindings fail to deprovision
		* the cluster is in an unhealthy state
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			if err := appsClient.DeleteInForeground(p.Namespace, appName); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleting app %q asynchronously... For progress on enabling this to run synchronously, see Kf Github issue #599.\n", appName)

			return nil
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
