// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apps

import (
	"fmt"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
)

// NewGetAppCommand creates a command to get details about a single application.
func NewGetAppCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var apps = &cobra.Command{
		Use:     "app APPNAME",
		Short:   "Get a pushed app",
		Example: `  kf app my-app`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Getting app %s in namespace: %s\n", appName, p.Namespace)

			app, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			var status, reason string
			if cond := app.Status.GetCondition("Ready"); cond != nil {
				status = fmt.Sprintf("%v", cond.Status)
				reason = cond.Reason
			}

			if !app.DeletionTimestamp.IsZero() {
				reason = "Deleting"
			}

			var host string
			url := app.Status.URL
			if url != nil {
				host = url.Host
			}

			fmt.Fprintln(w, "Name:", app.Name)
			fmt.Fprintln(w, "Space:", app.Namespace)
			fmt.Fprintln(w, "Status:", status)
			fmt.Fprintln(w, "Reason:", reason)
			fmt.Fprintln(w, "Host:", host)
			fmt.Fprintln(w)

			fmt.Fprintln(w, "Scale")
			instances := app.Spec.Instances
			fmt.Fprintln(w, "Stopped?:", instances.Stopped)
			if instances.Exactly != nil {
				fmt.Fprintln(w, "Instances:", instances.Exactly)
			}

			if instances.Min != nil {
				fmt.Fprintln(w, "Min:", instances.Min)
			}

			if instances.Max != nil {
				fmt.Fprintln(w, "Max:", instances.Max)
			}

			fmt.Fprintln(w, "Source")
			source := app.Spec.Source

			switch {
			case source.IsContainerBuild():
				containerImage := source.ContainerImage
				fmt.Fprintln(w, "Type: contaier")
				fmt.Fprintln(w, "Image:", containerImage.Image)
			case source.IsBuildpackBuild():
				buildpackBuild := source.BuildpackBuild
				fmt.Fprintln(w, "Type: buildpack")
				fmt.Fprintln(w, "Source:", buildpackBuild.Source)
				fmt.Fprintln(w, "Stack:", buildpackBuild.Stack)
				fmt.Fprintln(w, "Builder:", buildpackBuild.BuildpackBuilder)
				fmt.Fprintln(w, "Registry:", buildpackBuild.Registry)
				fmt.Fprintln(w, "Env:")
				for _, v := range buildpackBuild.Env {
					fmt.Fprintln(w, "-", v.Name, "=", v.Value)
				}
			}
			fmt.Fprintln(w)

			{
				fmt.Fprintln(w, "Runtime")
				// template := app.Spec.Template
				status := app.Status

				fmt.Fprintln(w, "Runtime Image:", status.Image)
				//
				// if len(template.Containers) >= 0 {
				// 	container := template.Containers[0]
				// 	fmt.Fprintln(w, "Env:")
				// 	for _, v := range container.Env {
				// 		fmt.Fprintln(w, "-", v.Name, "=", v.Value)
				// 	}
				// }

				fmt.Fprintln(w)
			}

			return nil
		},
	}

	return apps
}
