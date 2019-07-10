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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"github.com/spf13/cobra"
)

// NewUnmapRouteCommand creates a MapRoute command.
func NewUnmapRouteCommand(
	p *config.KfParams,
	c routes.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Unmap a route from an app",
		Example: `
  kf unmap-route myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route -n myspace myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			mutator := routes.Mutator(func(r *v1alpha1.Route) error {
				idx := -1
				for i, s := range r.Spec.KnativeServiceNames {
					if s == appName {
						idx = i
						break
					}
				}

				if idx < 0 {
					return fmt.Errorf("App %s not found", appName)
				}

				r.Spec.KnativeServiceNames = append(
					r.Spec.KnativeServiceNames[:idx],
					r.Spec.KnativeServiceNames[idx+1:]...,
				)
				return nil
			})

			urlPath = path.Join("/", urlPath)
			ksvcName := resources.VirtualServiceName(hostname, domain, urlPath)
			if err := c.Transform(p.Namespace, ksvcName, mutator); err != nil {
				return fmt.Errorf("failed to unmap Route: %s", err)
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
