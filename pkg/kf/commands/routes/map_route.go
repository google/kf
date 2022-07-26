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

// NewMapRouteCommand creates a MapRoute command.
func NewMapRouteCommand(
	p *config.KfParams,
	appsClient apps.Client,
) *cobra.Command {
	var (
		async        utils.AsyncFlags
		bindingFlags routeBindingFlags
		weight       int32
	)

	cmd := &cobra.Command{
		Use:   "map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [--weight WEIGHT]",
		Short: "Grant an App access to receive traffic from the Route.",
		Long: `
		Mapping an App to a Route will cause traffic to be forwarded to the App if
		the App has instances that are running and healthy.

		If multiple Apps are mapped to the same Route they will split traffic
		between them roughly evenly. Incoming network traffic is handled by multiple
		gateways which update their routing tables with slight delays and route
		independently. Because of this, traffic routing may not appear even but it
		will converge over time.
		`,
		Example: `
		kf map-route myapp example.com --hostname myapp # myapp.example.com
		kf map-route myapp myapp.example.com # myapp.example.com
		kf map-route myapp example.com --hostname myapp --weight 2 # myapp.example.com, myapp receives 2x traffic
		kf map-route --space myspace myapp example.com --hostname myapp # myapp.example.com
		kf map-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
		`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}
			appName, domain := args[0], args[1]

			mutator := func(app *v1alpha1.App) error {
				toAdd := bindingFlags.RouteWeightBinding(domain)
				toAdd.Weight = &weight

				apps.NewFromApp(app).MergeRoute(toAdd)
				return nil
			}

			if _, err := appsClient.Transform(cmd.Context(), p.Space, appName, mutator); err != nil {
				return fmt.Errorf("failed to map Route: %s", err)
			}

			action := fmt.Sprintf("Mapping route to app %q in space %q", appName, p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := appsClient.WaitForConditionRoutesReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)
	bindingFlags.Add(cmd)

	cmd.Flags().Int32Var(
		&weight,
		"weight",
		1,
		"Weight for the Route.",
	)

	return cmd
}
