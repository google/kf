// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicebindings

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/routes"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/spf13/cobra"
)

// NewUnbindRouteServiceCommand allows users to unbind a service instance from an HTTP route.
func NewUnbindRouteServiceCommand(p *config.KfParams, client serviceinstancebindings.Client) *cobra.Command {
	var (
		routeFlags routes.RouteFlags
		async      utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:     "unbind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE",
		Aliases: []string{"us"},
		Short:   "Unbind a route service instance from an HTTP route.",
		Long: `PREVIEW: this feature is not ready for production use.
		Unbinding a route service from an HTTP route causes requests to go straight to the
		HTTP route, rather than being processed by the route service first.
		`,
		Example: `kf unbind-route-service company.com --hostname myapp --path mypath myauthservice`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if p.FeatureFlags(ctx).RouteServices().IsDisabled() {
				return errors.New(`Route services feature is toggled off. Set "enable_route_services" to true in "config-defaults" to enable route services`)
			}

			domain := args[0]
			instanceName := args[1]
			bindingName := v1alpha1.MakeRouteServiceBindingName(routeFlags.Hostname, domain, routeFlags.Path, instanceName)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			if err := client.Delete(ctx, p.Space, bindingName); err != nil {
				return err
			}

			action := fmt.Sprintf("Deleting service instance binding in Space %q", p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForDeletion(context.Background(), p.Space, bindingName, 1*time.Second)
				if err != nil {
					return fmt.Errorf("unbind failed: %s", err)
				}
				return nil
			})
		},
		SilenceUsage: true,
	}

	async.Add(cmd)
	routeFlags.Add(cmd)

	return cmd
}
