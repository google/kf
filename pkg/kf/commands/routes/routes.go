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
	"strings"
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/spf13/cobra"
)

// NewRoutesCommand creates a Routes command.
func NewRoutesCommand(
	p *config.KfParams,
	c routes.Client,
	ac apps.Client,
) *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "List routes in space",
		Example: `
  kf routes
  `,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			fmt.Fprintf(cmd.OutOrStdout(), "Getting routes in namespace: %s\n", p.Namespace)

			routes, err := c.List(p.Namespace)
			if err != nil {
				return fmt.Errorf("failed to fetch Routes: %s", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d routes in namespace %s\n", len(routes), p.Namespace)
			fmt.Fprintln(cmd.OutOrStdout())

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 2, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "HOST\tDOMAIN\tPATH\tAPPS")
			for _, route := range routes {
				var apps []string
				for _, app := range route.Spec.KnativeServiceNames {

					ksvc, err := ac.Get(p.Namespace, app)
					if err != nil {
						return fmt.Errorf("fetching Knative Service failed: %s", err)
					}

					// TODO(poy): We might need to switch from the name to the
					// KfApp OwnerReference.
					apps = append(apps, ksvc.Name)
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", route.Spec.Hostname, route.Spec.Domain, route.Spec.Path, strings.Join(apps, ", "))
			}

			w.Flush()
			return nil
		},
	}
}

func splitHost(h string) (subDomain, domain string) {
	// A subdomain implies there are at least 2 periods. If parts has a length
	// less than 3, then we don't have a subdomain.
	parts := strings.SplitN(h, ".", 3)

	if len(parts) != 3 {
		return "", h
	}

	return parts[0], strings.Join(parts[1:], ".")
}
