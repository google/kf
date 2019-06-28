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
	"errors"
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateRouteCommand creates a CreateRoute command.
func NewCreateRouteCommand(
	p *config.KfParams,
	c routes.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "create-route DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Create a route",
		Example: `
  # Using namespace (instead of SPACE)
  kf create-route example.com --hostname myapp # myapp.example.com
  kf create-route -n myspace example.com --hostname myapp # myapp.example.com
  kf create-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath

  # [DEPRECATED] Using SPACE to match 'cf'
  kf create-route myspace example.com --hostname myapp # myapp.example.com
  kf create-route myspace example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			space, domain := p.Namespace, args[0]
			if len(args) == 2 {
				space = args[0]
				domain = args[1]
				fmt.Fprintln(cmd.OutOrStderr(), `
[WARN]: passing the SPACE as an argument is deprecated.
Instead use the --namespace flag.`)
			}

			if p.Namespace != "" && p.Namespace != space {
				return fmt.Errorf("SPACE (argument=%q) and namespace (flag=%q) (if provided) must match", space, p.Namespace)
			}

			if hostname == "" {
				return errors.New("--hostname is required")
			}

			cmd.SilenceUsage = true

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
					Hostname: hostname,
					Domain:   domain,
					Path:     urlPath,
				},
			}

			if _, err := c.Create(space, r); err != nil {
				return fmt.Errorf("failed to create Route: %s", err)
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
