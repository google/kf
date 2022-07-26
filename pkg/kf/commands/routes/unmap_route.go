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
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewUnmapRouteCommand creates a MapRoute command.
func NewUnmapRouteCommand(
	p *config.KfParams,
	appsClient apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags
	var bindingFlags routeBindingFlags

	cmd := &cobra.Command{
		Use:   "unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]",
		Short: "Revoke an App's access to receive traffic from the Route.",
		Long: `
		Unmapping an App from a Route will cause traffic matching the Route to no
		longer be forwarded to the App.

		The App may still receive traffic from an unmapped Route for a small period
		of time while the traffic rules on the gateways are propagated.

		The Route will re-balance its routing weights so other Apps mapped to it
		will receive the traffic. If no other Apps are bound the Route will return
		a 404 HTTP status code.
		`,
		Example: `
		# Unmap myapp.example.com from myapp in the targeted Space
		kf unmap-route myapp example.com --hostname myapp

		# Unmap the Route in a specific Space
		kf unmap-route --space myspace myapp example.com --hostname myapp

		# Unmap a Route with a path
		kf unmap-route myapp example.com --hostname myapp --path /mypath
		`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			binding := bindingFlags.RouteWeightBinding(domain)

			// Load the space so we can get the default domain
			space, err := p.GetTargetSpace(cmd.Context())
			if err != nil {
				return err
			}

			ctx := v1alpha1.WithRouteDefaultDomain(
				context.Background(),
				space.DefaultDomainOrBlank(),
			)

			if _, err := appsClient.Transform(cmd.Context(), p.Space, appName, func(app *v1alpha1.App) error {
				foundRoute := false
				for _, route := range app.Status.Routes {
					routeBinding := route.ToUnqualified()
					if routeBinding.EqualsBinding(ctx, binding) {
						foundRoute = true
						break
					}
				}

				if !foundRoute {
					return fmt.Errorf("Route %s not mapped to App %s", binding.String(), appName)
				}

				apps.NewFromApp(app).RemoveRoute(ctx, binding)
				return nil
			}); err != nil {
				return fmt.Errorf("failed to unmap Route: %s", err)
			}

			action := fmt.Sprintf("Unmapping Route to App %q in Space %q", appName, p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := appsClient.WaitForConditionRoutesReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)
	bindingFlags.Add(cmd)

	return cmd
}
