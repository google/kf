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

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/istio"
	"github.com/spf13/cobra"
)

// NewProxyRouteCommand creates a command capable of proxying a remote server locally.
func NewProxyRouteCommand(p *config.KfParams, ingressLister istio.IngressLister) *cobra.Command {
	var (
		gateway string
		port    int
		noStart bool
	)

	cmd := &cobra.Command{
		Use:     "proxy-route ROUTE",
		Short:   "Create a proxy to a route on a local port",
		Example: `kf proxy-route myhost.example.com`,
		Long: `
	This command creates a local proxy to a remote gateway modifying the request
	headers to make requests with the host set as the specified route.

	You can manually specify the gateway or have it autodetected based on your
	cluster.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			routeHost := args[0]
			cmd.SilenceUsage = true

			if gateway == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Autodetecting app gateway. Specify a custom gateway using the --gateway flag.")

				ingress, err := istio.ExtractIngressFromList(ingressLister.ListIngresses())
				if err != nil {
					return err
				}
				gateway = ingress
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			if noStart {
				fmt.Fprintln(w, "exiting proxy because no-start flag was provided")
				utils.PrintCurlExamples(w, listener, routeHost, gateway, false)
				return nil
			}

			utils.PrintCurlExamples(w, listener, routeHost, gateway, true)
			return http.Serve(listener, utils.CreateProxy(cmd.OutOrStdout(), routeHost, gateway))
		},
	}

	cmd.Flags().StringVar(
		&gateway,
		"gateway",
		"",
		"HTTP gateway to route requests to (default: autodetected from cluster)",
	)

	cmd.Flags().IntVar(
		&port,
		"port",
		8080,
		"Local port to listen on",
	)

	cmd.Flags().BoolVar(
		&noStart,
		"no-start",
		false,
		"Exit before starting the proxy",
	)
	cmd.Flags().MarkHidden("no-start")

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
