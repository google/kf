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
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/routeclaims"
	"github.com/spf13/cobra"
)

// NewDeleteRouteCommand creates a DeleteRoute command. To delete a route and
// not have the controller bring it back, delete is a multistep process.
// First, unmap the route from each app that has it. Next, delete the
// RouteClaim.
func NewDeleteRouteCommand(
	p *config.KfParams,
	c routeclaims.Client,
	a apps.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "delete-route DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Delete a route",
		Example: `
  kf delete-route example.com --hostname myapp # myapp.example.com
  kf delete-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			domain := args[0]
			cmd.SilenceUsage = true

			route := v1alpha1.RouteSpecFields{
				Hostname: hostname,
				Domain:   domain,
				Path:     urlPath,
			}

			// TODO: This is O(apps). We could do better if we lookup the
			// routes with the proper labels first, then use their apps.
			apps, err := a.List(
				p.Namespace,
				apps.WithListFilters([]apps.Predicate{
					func(app *v1alpha1.App) bool {
						return algorithms.Search(
							0,
							v1alpha1.RouteSpecFieldsSlice{route},
							v1alpha1.RouteSpecFieldsSlice(app.Spec.Routes),
						)
					},
				}),
			)
			if err != nil {
				return fmt.Errorf("failed to list apps: %s", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleting route asynchronously... For progress on enabling this to run synchronously, see Kf Github issue #599.\n")

			for _, app := range apps {
				fmt.Fprintf(cmd.OutOrStderr(), "Unmapping route from %s...\n", app.Name)
				if err := unmapApp(
					p.Namespace,
					app.Name,
					route,
					a,
					cmd,
				); err != nil {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStderr(), "Deleting route claim...\n")
			if err := c.Delete(
				p.Namespace,
				v1alpha1.GenerateRouteClaimName(
					hostname,
					domain,
					urlPath,
				),
			); err != nil {
				return fmt.Errorf("failed to delete Route: %s", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(
		&hostname,
		"hostname",
		"",
		"Hostname for the route",
	)
	cmd.Flags().StringVar(
		&urlPath,
		"path",
		"",
		"URL Path for the route",
	)

	return cmd
}
