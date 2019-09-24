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
	"io"
	"path"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewUnmapRouteCommand creates a MapRoute command.
func NewUnmapRouteCommand(
	p *config.KfParams,
	c apps.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Unmap a route from an app",
		Example: `
  kf unmap-route myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route --namespace myspace myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			route := v1alpha1.RouteSpecFields{
				Hostname: hostname,
				Domain:   domain,
				Path:     path.Join("/", urlPath),
			}

			return unmapApp(p.Namespace, appName, route, c, cmd.OutOrStdout())
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

func unmapApp(
	namespace string,
	appName string,
	route v1alpha1.RouteSpecFields,
	c apps.Client,
	w io.Writer,
) error {
	mutator := apps.Mutator(func(app *v1alpha1.App) error {
		// Ensure the App has the Route, if not return an error.
		if !algorithms.Search(
			0,
			v1alpha1.RouteSpecFieldsSlice{route},
			v1alpha1.RouteSpecFieldsSlice(app.Spec.Routes),
		) {
			return fmt.Errorf("App %s not found", app.Name)
		}

		app.Spec.Routes = []v1alpha1.RouteSpecFields(
			(algorithms.Delete(
				v1alpha1.RouteSpecFieldsSlice(app.Spec.Routes),
				v1alpha1.RouteSpecFieldsSlice{route},
			)).(v1alpha1.RouteSpecFieldsSlice))
		return nil
	})

	if _, err := c.Transform(namespace, appName, mutator); err != nil {
		return fmt.Errorf("failed to unmap Route: %s", err)
	}

	fmt.Fprintf(w, "Unmapping route... %s", utils.AsyncLogSuffix)
	return nil
}
