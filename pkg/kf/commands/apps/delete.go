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

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates a delete command.
func NewDeleteCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var async utils.AsyncFlags

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

		Apps may take a long time to delete if:

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

			if err := appsClient.Delete(p.Namespace, appName); err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.OutOrStdout(), fmt.Sprintf("Deleting app %s", appName), func() error {
				if _, err := appsClient.WaitForDeletion(context.Background(), p.Namespace, appName, 1*time.Second); err != nil {
					return fmt.Errorf("couldn't delete: %s", err)
				}

				return nil
			})
		},
	}

	async.Add(cmd)

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
