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

package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/routes"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewDeleteRouteCommand creates a DeleteRoute command. To delete a route and
// not have the controller bring it back, delete is a multistep process.
// First, unmap the route from each app that has it. Next, delete the
// Route.
func NewDeleteRouteCommand(
	p *config.KfParams,
	c routes.Client,
	a apps.Client,
) *cobra.Command {
	var (
		routeFlags RouteFlags
		async      utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:   "delete-route DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Delete a Route in the targeted Space.",
		Example: `
  # Delete the Route myapp.example.com
  kf delete-route example.com --hostname myapp
  # Delete a Route on a path myapp.example.com/mypath
  kf delete-route example.com --hostname myapp --path /mypath
  `,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			domain := args[0]

			route := routeFlags.RouteSpecFields(domain)

			// TODO: This is O(apps). We could do better if we lookup the
			// routes with the proper labels first, then use their apps.
			spaceApps, err := a.List(ctx, p.Space)
			if err != nil {
				return fmt.Errorf("failed to list Apps: %s", err)
			}

			logger.Info("Unmapping bound Apps...")

			for _, app := range spaceApps {
				if !apps.NewFromApp(&app).HasMatchingRoutes(route) {
					continue
				}

				logger.Infof("Unmapping Route from %s...", app.Name)

				if _, err := a.Transform(ctx, p.Space, app.Name, func(app *v1alpha1.App) error {
					apps.NewFromApp(app).RemoveRoutesForClaim(route)
					return nil
				}); err != nil {
					return fmt.Errorf("failed to unmap Route: %s", err)
				}
			}

			instanceName := v1alpha1.GenerateRouteName(
				routeFlags.Hostname,
				domain,
				routeFlags.Path,
			)

			action := fmt.Sprintf("Deleting Route %q in Space %q", instanceName, p.Space)

			logger.Info("Deleting Route...")
			if err := c.Delete(ctx, p.Space, instanceName); err != nil {
				return fmt.Errorf("failed to delete Route: %s", err)
			}

			return async.AwaitAndLog(cmd.OutOrStderr(), action, func() error {
				_, err := c.WaitForDeletion(context.Background(), p.Space, instanceName, 1*time.Second)
				return err
			})
		},
	}
	async.Add(cmd)
	routeFlags.Add(cmd)

	return cmd
}
