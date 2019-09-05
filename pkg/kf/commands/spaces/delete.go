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
	"fmt"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

// NewDeleteSpaceCommand allows users to delete spaces.
func NewDeleteSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
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

		NOTE: Space deletion is asynchronous and may take a long time to complete
		depending on the number of items in the space.

		You will be unable to make changes to resources in the space once deletion
		has begun.
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "Deleting space %s asynchronously. For progress on enabling this to run synchronously, see Kf Github issue #599.\n", name)
			return client.Delete(name)
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}
