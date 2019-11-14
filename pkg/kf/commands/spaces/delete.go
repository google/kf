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

package spaces

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

// NewDeleteSpaceCommand allows users to delete spaces.
func NewDeleteSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:     "delete-space SPACE",
		Short:   "Delete a space",
		Example: `kf delete-space my-space`,
		Long: `Delete a space and all its contents.

		This will delete a space's:

		* Apps
		* Service bindings
		* Service instances
		* RBAC roles
		* Routes
		* The backing Kubernetes namespace
		* Anything else in that namespace

		You will be unable to make changes to resources in the space once deletion
		has begun.
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			if err := client.Delete(name); err != nil {
				return fmt.Errorf("failed to delete space: %s", err)
			}

			action := fmt.Sprintf("Deleting space %q", name)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForDeletion(context.Background(), name, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}
