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

package apps

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/reconciler/route/resources"
	"github.com/google/kf/v2/pkg/system"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewProxyCommand creates a command capable of proxying a remote server locally.
func NewProxyCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var (
		gateway string
		port    int
		noStart bool
	)

	cmd := &cobra.Command{
		Use:     "proxy APP_NAME",
		Short:   "Start a local reverse proxy to an App.",
		Example: `kf proxy myapp`,
		Long: `
		Proxy creates a reverse HTTP proxy to the cluster's gateway on a local
		port opened on the operating system's loopback device.

		The proxy rewrites all HTTP requests, changing the HTTP Host header
		and adding an additional header X-Kf-App to ensure traffic reaches
		the specified App even if multiple are attached to the same route.

		Proxy does not establish a direct connection to the App.

		For proxy to work:

		* The cluster's gateway must be accessible from your local machine.
		* The App must have a public URL

		If you need to establish a direct connection to an App, use the
		port-forward command in kubectl. It establishes a proxied connection
		directly to a port on a pod via the Kubernetes cluster. port-forward
		bypasses all routing.
		`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			logger := logging.FromContext(ctx)

			app, err := appsClient.Get(ctx, p.Space, appName)
			if err != nil {
				return err
			}

			routes := app.Status.Routes
			if len(routes) == 0 {
				return fmt.Errorf("no public routes for App %s", appName)
			}

			appDomain := ""
			for _, route := range routes {
				// we can't use wildcard DNS entries because we won't know the
				// valid host
				if route.Source.IsWildcard() {
					continue
				}

				appDomain = route.Source.Host()
				break
			}

			if appDomain == "" {
				return errors.New("couldn't find suitable App domain")
			}

			if gateway == "" {
				logger.Info("Autodetecting App gateway. Specify a custom gateway using the --gateway flag.")

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
				logger.Info("Exiting because no-start flag was provided")
				utils.PrintCurlExamplesNoListener(ctx, appDomain, gateway)
				return nil
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				return err
			}

			utils.PrintCurlExamples(ctx, listener, appDomain, gateway)

			// Write the address to stdout so it can be consumed by scripts.
			fmt.Fprintln(cmd.OutOrStdout(), "http://"+listener.Addr().String())

			additional := http.Header{
				resources.KfAppMatchHeader: []string{appName},
			}

			return http.Serve(listener, utils.CreateProxy(cmd.OutOrStderr(), appDomain, gateway, additional))
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
