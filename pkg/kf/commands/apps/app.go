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
	"io"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/spf13/cobra"
)

// NewGetAppCommand creates a command to get details about a single application.
func NewGetAppCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "app APP_NAME",
		Short:   "Print information about a deployed app",
		Long:    `Prints information about a deployed app.`,
		Example: `kf app my-app`,
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

			describe.ObjectMeta(w, app.ObjectMeta)
			fmt.Fprintln(w)

			describe.DuckStatus(w, app.Status.Status)
			fmt.Fprintln(w)

			describe.AppSpecInstances(w, app.Spec.Instances)
			fmt.Fprintln(w)

			describe.AppSpecTemplate(w, app.Spec.Template)
			fmt.Fprintln(w)

			describe.SourceSpec(w, app.Spec.Source)
			fmt.Fprintln(w)

			describe.SectionWriter(w, "Runtime", func(w io.Writer) {
				status := app.Status

				fmt.Fprintf(w, "Image:\t%s\n", status.Image)
				if url := status.URL; url != nil {
					fmt.Fprintf(w, "Host:\t%s\n", url.Host)
				}

				kfApp := apps.NewFromApp(app)
				describe.HealthCheck(w, kfApp.GetHealthCheck())
				describe.EnvVars(w, kfApp.GetEnvVars())
			})
			fmt.Fprintln(w)

			return nil
		},
	}

	completion.MarkArgCompletionSupported(cmd, "apps")

	return cmd
}
