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
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/logging"
)

// NewBindRouteServiceCommand allows users to bind a route service to a route.
func NewBindRouteServiceCommand(p *config.KfParams, client serviceinstancebindings.Client, secretsClient secrets.Client) *cobra.Command {
	var (
		routeFlags   routes.RouteFlags
		configAsJSON string
		async        utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:     "bind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]",
		Aliases: []string{"brs"},
		Short:   "Bind a route service instance to an HTTP route.",
		Long: `PREVIEW: this feature is not ready for production use.
		Binding a service to an HTTP route causes traffic to be processed by that service before the requests are forwarded to the route.
		`,
		Example:      `  kf bind-route-service company.com --hostname myapp --path mypath myauthservice`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
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

			paramBytes, err := utils.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			paramsSecretName := v1alpha1.MakeRouteServiceBindingParamsSecretName(routeFlags.Hostname, domain, routeFlags.Path, instanceName)

			desiredBinding := &v1alpha1.ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      bindingName,
					Namespace: p.Space,
				},
				Spec: v1alpha1.ServiceInstanceBindingSpec{
					BindingType: v1alpha1.BindingType{
						Route: &v1alpha1.RouteRef{
							Hostname: routeFlags.Hostname,
							Domain:   domain,
							Path:     routeFlags.Path,
						},
					},
					InstanceRef: v1.LocalObjectReference{
						Name: instanceName,
					},
					ParametersFrom: v1.LocalObjectReference{
						Name: paramsSecretName,
					},
				},
			}

			logging.FromContext(ctx).Infof("Creating service instance binding %q in space %q", bindingName, p.Space)

			actualBinding, err := client.Create(ctx, p.Space, desiredBinding)
			if err != nil {
				return err
			}

			if _, err := secretsClient.CreateParamsSecret(ctx, actualBinding, paramsSecretName, paramBytes); err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for service instance binding to become ready", func() (err error) {
				_, err = client.WaitForConditionReadyTrue(context.Background(), p.Space, bindingName, 1*time.Second)
				if err != nil {
					return fmt.Errorf("bind failed: %s", err)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(
		&configAsJSON,
		"parameters",
		"c",
		"{}",
		"JSON object or path to a JSON file containing configuration parameters.")

	async.Add(cmd)
	routeFlags.Add(cmd)

	return cmd
}
