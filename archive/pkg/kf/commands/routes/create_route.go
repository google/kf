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
	"path"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/routeclaims"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateRouteCommand creates a CreateRoute command.
func NewCreateRouteCommand(
	p *config.KfParams,
	c routeclaims.Client,
) *cobra.Command {
	var hostname, urlPath string

	cmd := &cobra.Command{
		Use:   "create-route DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Create a route",
		Example: `
  # Using namespace (instead of SPACE)
  kf create-route example.com --hostname myapp # myapp.example.com
  kf create-route --namespace myspace example.com --hostname myapp # myapp.example.com
  kf create-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath

  # [DEPRECATED] Using SPACE to match 'cf'
  kf create-route myspace example.com --hostname myapp # myapp.example.com
  kf create-route myspace example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  `,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			space, domain := p.Namespace, args[0]
			if len(args) == 2 {
				space = args[0]
				domain = args[1]
				fmt.Fprintln(cmd.OutOrStderr(), `
[WARN]: passing the SPACE as an argument is deprecated.
Use the --namespace flag instead.`)
			}

			if p.Namespace != "" && p.Namespace != "default" && p.Namespace != space {
				return fmt.Errorf("SPACE (argument=%q) and namespace (flag=%q) (if provided) must match", space, p.Namespace)
			}

			if hostname == "" {
				return errors.New("--hostname is required")
			}

			cmd.SilenceUsage = true

			urlPath = path.Join("/", urlPath)

			r := &v1alpha1.RouteClaim{
				TypeMeta: metav1.TypeMeta{
					Kind: "RouteClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: space,
					Name: v1alpha1.GenerateRouteClaimName(
						hostname,
						domain,
						urlPath,
					),
				},
				Spec: v1alpha1.RouteClaimSpec{
					RouteSpecFields: v1alpha1.RouteSpecFields{
						Hostname: hostname,
						Domain:   domain,
						Path:     urlPath,
					},
				},
			}

			if _, err := c.Create(space, r); err != nil {
				return fmt.Errorf("failed to create Route: %s", err)
			}

			// NOTE: RouteClaims don't have a status so there's nothing to wait on
			// after creation.
			fmt.Fprintln(cmd.OutOrStdout(), "Creating route...")
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
