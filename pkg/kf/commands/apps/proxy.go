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
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
)

// NewProxyCommand creates a command capable of proxying a remote server locally.
func NewProxyCommand(p *config.KfParams, appsClient apps.Client, ingressLister kf.IngressLister) *cobra.Command {
	var (
		gateway string
		port    int
		noStart bool
	)

	var proxy = &cobra.Command{
		Use:     "proxy APP_NAME",
		Short:   "Creates a proxy to an app on a local port.",
		Example: `  kf proxy myapp`,
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

			if gateway == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Autodetecting app gateway. Specify a custom gateway using the --gateway flag.")

				ingress, err := kf.ExtractIngressFromList(ingressLister.ListIngresses())
				if err != nil {
					return err
				}
				gateway = ingress
			}

			app, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Forwarding requests from http://%s to http://%s\n", listener.Addr(), gateway)

			if noStart {
				fmt.Fprintln(cmd.OutOrStdout(), "exiting because no-start flag was provided")
				return nil
			}

			return http.Serve(listener, createProxy(cmd.OutOrStdout(), app.Status.URL.Host, gateway))
		},
	}

	proxy.Flags().StringVar(
		&gateway,
		"gateway",
		"",
		"the HTTP gateway to route requests to, if unset it will be autodetected",
	)

	proxy.Flags().IntVar(
		&port,
		"port",
		8080,
		"the local port to attach to",
	)

	proxy.Flags().BoolVar(
		&noStart,
		"no-start",
		false,
		"don't actually start the HTTP proxy",
	)
	proxy.Flags().MarkHidden("no-start")

	return proxy
}

func createProxy(w io.Writer, appHost, gateway string) *httputil.ReverseProxy {
	logger := log.New(w, fmt.Sprintf("\033[34m[%s via %s]\033[0m ", appHost, gateway), log.Ltime)
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = appHost
			req.URL.Scheme = "http"
			req.URL.Host = gateway

			logger.Printf("%s %s\n", req.Method, req.URL.RequestURI())
		},
		ErrorLog: logger,
	}
}
