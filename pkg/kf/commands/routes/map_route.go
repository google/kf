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
	"path"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
)

// NewMapRouteCommand creates a MapRoute command.
func NewMapRouteCommand(
	p *config.KfParams,
	appsClient apps.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Map a route to an app",
		Example: `
  kf map-route myapp example.com --hostname myapp # myapp.example.com
  kf map-route --namespace myspace myapp example.com --hostname myapp # myapp.example.com
  kf map-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Routes = append(
					app.Spec.Routes,
					v1alpha1.RouteSpecFields{
						Hostname: hostname,
						Domain:   domain,
						Path:     path.Join("/", urlPath),
					},
				)
				return nil
			}

			err := appsClient.Transform(p.Namespace, appName, mutator)
			if err != nil {
				return fmt.Errorf("failed to map Route: %s", err)
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
