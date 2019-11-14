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
	"fmt"
	"net"
	"net/http"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/istio"
	"github.com/spf13/cobra"
)

// NewProxyCommand creates a command capable of proxying a remote server locally.
func NewProxyCommand(p *config.KfParams, appsClient apps.Client, ingressLister istio.IngressLister) *cobra.Command {
	var (
		gateway string
		port    int
		noStart bool
	)

	cmd := &cobra.Command{
		Use:     "proxy APP_NAME",
		Short:   "Create a proxy to an app on a local port",
		Example: `kf proxy myapp`,
		Long: `
	This command creates a local proxy to a remote gateway modifying the request
	headers to make requests route to your app.

	You can manually specify the gateway or have it autodetected based on your
	cluster.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			cmd.SilenceUsage = true

			app, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			url := app.Status.URL
			if url == nil {
				return fmt.Errorf("No route for app %s", appName)
			}

			if gateway == "" {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Autodetecting app gateway. Specify a custom gateway using the --gateway flag."); err != nil {
					return err
				}

				ingress, err := istio.ExtractIngressFromList(ingressLister.ListIngresses())
				if err != nil {
					return err
				}
				gateway = ingress
			}

			appHost := url.Host
			w := cmd.OutOrStdout()

			if noStart {
				if _, err := fmt.Fprintln(w, "exiting because no-start flag was provided"); err != nil {
					return err
				}
				if err := utils.PrintCurlExamplesNoListener(w, appHost, gateway); err != nil {
					return err
				}
				return nil
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				return err
			}

			if err := utils.PrintCurlExamples(w, listener, appHost, gateway); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, "\033[33mNOTE: the first request may take some time if the app is scaled to zero\033[0m"); err != nil {
				return err
			}

			return http.Serve(listener, utils.CreateProxy(cmd.OutOrStdout(), app.Status.URL.Host, gateway))
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
	if err := cmd.Flags().MarkHidden("no-start"); err != nil {
		panic(err)
	}

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
