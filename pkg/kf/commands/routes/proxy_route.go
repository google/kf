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
	"net"
	"net/http"

	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/system"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewProxyRouteCommand creates a command capable of proxying a remote server locally.
func NewProxyRouteCommand(p *config.KfParams) *cobra.Command {
	var (
		gateway string
		port    int
		noStart bool
	)

	cmd := &cobra.Command{
		Use:     "proxy-route ROUTE",
		Short:   "Start a local reverse proxy to a Route.",
		Example: `kf proxy-route myhost.example.com`,
		Long: `
		Proxy route creates a reverse HTTP proxy to the cluster's gateway on a local
		port opened on the operating system's loopback device.

		The proxy rewrites all HTTP requests, changing the HTTP Host header to match
		the Route. If multiple Apps are mapped to the same route, the traffic sent
		over the proxy will follow their routing rules with regards to weight.
		If no Apps are mapped to the route, traffic sent through the proxy will
		return a HTTP 404 status code.

		Proxy route DOES NOT establish a direct connection to any Kubernetes resource.

		For proxy to work:

		* The cluster's gateway MUST be accessible from your local machine.
		* The Route MUST have a public URL
		`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			routeHost := args[0]

			if gateway == "" {
				logger.Info("Autodetecting app gateway. Specify a custom gateway using the --gateway flag.")

				space, err := p.GetTargetSpace(ctx)
				if err != nil {
					return err
				}

				ingress, err := system.ExtractProxyIngressFromList(space.Status.IngressGateways)
				if err != nil {
					return err
				}
				gateway = ingress
			}

			if noStart {
				logger.Info("exiting proxy because no-start flag was provided")
				utils.PrintCurlExamplesNoListener(ctx, routeHost, gateway)
				return nil
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				return err
			}

			utils.PrintCurlExamples(ctx, listener, routeHost, gateway)

			// Write the address to stdout so it can be consumed by scripts.
			fmt.Fprintln(cmd.OutOrStdout(), "http://"+listener.Addr().String())

			return http.Serve(listener, utils.CreateProxy(cmd.OutOrStderr(), routeHost, gateway, http.Header{}))
		},
	}

	cmd.Flags().StringVar(
		&gateway,
		"gateway",
		"",
		"IP address of the HTTP gateway to route requests to.",
	)

	cmd.Flags().IntVar(
		&port,
		"port",
		8080,
		"Local port to listen on.",
	)

	cmd.Flags().BoolVar(
		&noStart,
		"no-start",
		false,
		"Exit before starting the proxy.",
	)
	cmd.Flags().MarkHidden("no-start")

	return cmd
}
