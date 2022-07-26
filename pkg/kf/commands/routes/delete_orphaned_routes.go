// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routes

import (
	"fmt"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/routes"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewDeleteOrphanedRoutesCommand creates a command to delete orphaned routes.
func NewDeleteOrphanedRoutesCommand(
	p *config.KfParams,
	c routes.Client,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-orphaned-routes",
		Short: "Delete Routes with no App bindings.",
		Long: `Deletes Routes in the targeted Space that are fully reconciled and
    don't have any bindings to Apps.`,
		Example:      `kf delete-orphaned-routes`,
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			logger.Infof("Deleting orphaned Routes in Space: %s", p.Space)

			routes, err := c.List(ctx, p.Space)
			if err != nil {
				return fmt.Errorf("failed to list Routes: %v", err)
			}

			var numDeleted int
			for _, r := range routes {
				if r.IsOrphaned() {
					url := r.Spec.ToURL()
					logger.Infof("Deleting %s", url.String())
					numDeleted++

					if err := c.Delete(ctx, p.Space, r.Name); err != nil {
						return fmt.Errorf("failed to delete Route: %v", err)
					}
				}
			}

			logger.Infof("Deleted %d route(s)", numDeleted)
			return nil
		},
	}

	// Set up a mock force flag so users transitioning from cf won't have to
	// change scripts.
	cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation.")
	cmd.Flags().MarkHidden("force")

	return cmd
}
