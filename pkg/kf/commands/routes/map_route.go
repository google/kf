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

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewMapRouteCommand creates a MapRoute command.
func NewMapRouteCommand(
	p *config.KfParams,
	routesClient routes.Client,
	appsClient apps.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Map a route to an app",
		Example: `
  kf map-route myapp example.com --hostname myapp # myapp.example.com
  kf map-route -n myspace myapp example.com --hostname myapp # myapp.example.com
  kf map-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			_, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return fmt.Errorf("failed to fetch app: %s", err)
			}

			merger := routes.Merger(func(newR, oldR *v1alpha1.Route) *v1alpha1.Route {
				newR.ObjectMeta = *oldR.ObjectMeta.DeepCopy()

				// Ensure the app isn't already there
				for _, name := range oldR.Spec.KnativeServiceNames {
					if name == appName {
						continue
					}
					newR.Spec.KnativeServiceNames = append(newR.Spec.KnativeServiceNames, name)
				}

				return newR
			})

			urlPath = path.Join("/", urlPath)
			r := &v1alpha1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind: "Route",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: resources.VirtualServiceName(
						hostname,
						domain,
						urlPath,
					),
				},
				Spec: v1alpha1.RouteSpec{
					Hostname:            hostname,
					Domain:              domain,
					Path:                urlPath,
					KnativeServiceNames: []string{appName},
				},
			}

			if _, err := routesClient.Upsert(p.Namespace, r, merger); err != nil {
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
