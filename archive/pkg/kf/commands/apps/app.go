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
	"sort"
	"strings"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewGetAppCommand creates a command to get details about a single application.
func NewGetAppCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	printFlags := genericclioptions.NewPrintFlags("")

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

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			fmt.Fprintf(cmd.ErrOrStderr(), "Getting app %s in namespace: %s\n", appName, p.Namespace)

			app, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			if printFlags.OutputFlagSpecified() {
				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}

				// If the type didn't come back with a kind, update it with the
				// type we deserialized it with so the printer will work.
				app.GetObjectKind().SetGroupVersionKind(app.GetGroupVersionKind())
				return printer.PrintObj(app, w)
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

				kfApp := apps.NewFromApp(app)
				fmt.Fprintf(w, "Cluster URL\t%s\n", kfApp.GetClusterURL())

				describe.HealthCheck(w, kfApp.GetHealthCheck())
				describe.EnvVars(w, kfApp.GetEnvVars())
				describe.RouteSpecFieldsList(w, app.Spec.Routes)
			})
			fmt.Fprintln(w)

			return nil
		},
	}

	printFlags.AddFlags(cmd)

	// Override output format to be sorted so our generated documents are deterministic
	{
		allowedFormats := printFlags.AllowedFormats()
		sort.Strings(allowedFormats)
		cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(allowedFormats, "|"))
	}

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
