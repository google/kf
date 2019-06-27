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
	"net/http"
	"path"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/routeutil"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
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
				return errors.New("SPACE (argument) and namespace (if provided) must match")
			}

			if hostname == "" {
				return errors.New("--hostname is required")
			}

			cmd.SilenceUsage = true

			var pathMatchers []networking.HTTPMatchRequest
			if urlPath != "" {
				urlPath = path.Join("/", urlPath)
				pathMatchers = append(pathMatchers, networking.HTTPMatchRequest{
					URI: &v1alpha1.StringMatch{
						Prefix: urlPath,
					},
				})
			}

			hostDomain := hostname + "." + domain

			vs := &networking.VirtualService{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: routeutil.EncodeRouteName(hostname, domain, urlPath),
					Annotations: map[string]string{
						"domain":   domain,
						"hostname": hostname,
						"path":     urlPath,
					},
				},
				Spec: networking.VirtualServiceSpec{
					// TODO: Is this a constant?
					Gateways: []string{"knative-ingress-gateway.knative-serving.svc.cluster.local"},
					Hosts:    []string{hostDomain},
					HTTP: []networking.HTTPRoute{
						{
							Match: pathMatchers,
							Route: []networking.HTTPRouteDestination{
								{
									Destination: networking.Destination{
										// TODO: Is this a constant?
										Host: "istio-ingressgateway.istio-system.svc.cluster.local",

										// XXX: If this is not included, then
										// we get an error back from the
										// server suggesting we have to have a
										// port set. It doesn't seem to hurt
										// anything as we just return a fault.
										Port: networking.PortSelector{
											Number: 80,
										},
									},
									Weight: 100,
								},
							},
							Fault: &networking.HTTPFaultInjection{
								Abort: &networking.InjectAbort{
									Percent:    100,
									HTTPStatus: http.StatusServiceUnavailable,
								},
							},
						},
					},
				},
			}

			if _, err := c.Create(space, vs); err != nil {
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
