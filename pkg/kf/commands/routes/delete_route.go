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

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/routeutil"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/spf13/cobra"
)

// NewDeleteRouteCommand creates a DeleteRoute command.
func NewDeleteRouteCommand(
	p *config.KfParams,
	c routes.Client,
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
			domain := args[0]
			cmd.SilenceUsage = true
			urlPath = path.Join("/", urlPath)

			if err := c.Delete(p.Namespace, routeutil.EncodeRouteName(hostname, domain, urlPath)); err != nil {
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
