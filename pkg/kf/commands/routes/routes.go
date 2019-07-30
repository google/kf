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
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/routeclaims"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/spf13/cobra"
)

// NewRoutesCommand creates a Routes command.
func NewRoutesCommand(
	p *config.KfParams,
	r routes.Client,
	c routeclaims.Client,
	a apps.Client,
) *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "List routes in space",
		Example: `
  kf routes
  `,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			fmt.Fprintf(cmd.OutOrStdout(), "Getting routes in space: %s\n", p.Namespace)
			fmt.Fprintln(cmd.OutOrStdout())

			routes, err := r.List(p.Namespace)
			if err != nil {
				return fmt.Errorf("failed to fetch Routes: %s", err)
			}

			routeClaims, err := c.List(p.Namespace)
			if err != nil {
				return fmt.Errorf("failed to fetch RouteClaims: %s", err)
			}

			apps, err := a.List(p.Namespace)
			if err != nil {
				return fmt.Errorf("failed to fetch Apps: %s", err)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 2, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "host\tdomain\tpath\tapps")
			for _, route := range groupRoutes(routes, routeClaims) {
				names := strings.Join(appNames(apps, route), ", ")
				fmt.Fprintf(
					w,
					"%s\t%s\t%s\t%s\n",
					route.Hostname,
					route.Domain,
					route.Path,
					names,
				)
			}

			return w.Flush()
		},
	}
}

func groupRoutes(routes []v1alpha1.Route, claims []v1alpha1.RouteClaim) []v1alpha1.RouteSpecFields {
	var fields v1alpha1.RouteSpecFieldsSlice
	for _, r := range routes {
		fields = append(fields, r.Spec.RouteSpecFields)
	}
	for _, c := range claims {
		fields = append(fields, c.Spec.RouteSpecFields)
	}

	fields = algorithms.Dedupe(
		v1alpha1.RouteSpecFieldsSlice(fields),
	).(v1alpha1.RouteSpecFieldsSlice)
	sort.Sort(fields)

	return []v1alpha1.RouteSpecFields(fields)
}

func appNames(apps []v1alpha1.App, route v1alpha1.RouteSpecFields) []string {
	var names []string
	for _, app := range apps {
		// Look to see if App already has Route
		if !algorithms.Search(
			0,
			v1alpha1.RouteSpecFieldsSlice{route},
			v1alpha1.RouteSpecFieldsSlice(app.Spec.Routes),
		) {
			continue
		}

		names = append(names, app.Name)
	}
	return names
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
