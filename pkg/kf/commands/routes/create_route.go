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
	"context"
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/routes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/logging"
)

// NewCreateRouteCommand creates a CreateRoute command.
func NewCreateRouteCommand(
	p *config.KfParams,
	c routes.Client,
) *cobra.Command {
	var (
		routeFlags RouteFlags
		async      utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:   "create-route DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Create a traffic routing rule for a host+path pair.",
		Long: `
		Creating a Route allows Apps to declare they want to receive traffic on
		the same host/domain/path combination.

		Routes without any bound Apps (or with only stopped Apps) will return a 404
		HTTP status code.

		Kf doesn't enforce Route uniqueness between Spaces. It's recommended
		to provide each Space with its own subdomain instead.
		`,
		Example: `
		kf create-route example.com --hostname myapp # myapp.example.com
		kf create-route --space myspace example.com --hostname myapp # myapp.example.com
		kf create-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath
		kf create-route --space myspace myapp.example.com # myapp.example.com

		# Using SPACE to match 'cf'
		kf create-route myspace example.com --hostname myapp # myapp.example.com
		kf create-route myspace example.com --hostname myapp --path /mypath # myapp.example.com/mypath
		`,
		Args:         cobra.RangeArgs(1, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			space, domain := p.Space, args[0]
			if len(args) == 2 {
				space = args[0]
				domain = args[1]
			}

			if p.Space != "" && p.Space != "default" && p.Space != space {
				return fmt.Errorf("SPACE (argument=%q) and space (flag=%q) (if provided) must match", space, p.Space)
			}

			fields := routeFlags.RouteSpecFields(domain)

			instanceName := v1alpha1.GenerateRouteName(
				fields.Hostname,
				fields.Domain,
				fields.Path,
			)

			r := &v1alpha1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind: "Route",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: space,
					Name:      instanceName,
				},
				Spec: v1alpha1.RouteSpec{
					RouteSpecFields: fields,
				},
			}

			if _, err := c.Create(ctx, space, r); err != nil {
				return fmt.Errorf("failed to create Route: %s", err)
			}

			logging.FromContext(ctx).Infof("Creating Route %q in space %q", instanceName, p.Space)
			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for Route to become ready", func() (err error) {
				_, err = c.WaitForConditionReadyTrue(context.Background(), space, instanceName, 1*time.Second)
				return
			})
		},
	}
	async.Add(cmd)
	routeFlags.Add(cmd)

	return cmd
}
