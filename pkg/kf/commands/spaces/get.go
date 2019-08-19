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

package spaces

import (
	"fmt"
	"io"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/spaces"

	"github.com/spf13/cobra"
)

// NewGetSpaceCommand allows users to create spaces.
func NewGetSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space SPACE",
		Short: "Show space info",
		Long: `Get detailed information about a specific space and its configuration.

		The output of this command is similar to what you'd get by running:

		    kubectl describe space.kf.dev SPACE

		`,
		Example: `kf space my-space`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			space, err := client.Get(name)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			describe.ObjectMeta(w, space.ObjectMeta)
			fmt.Fprintln(w)

			describe.DuckStatus(w, space.Status.Status)
			fmt.Fprintln(w)

			describe.SectionWriter(w, "Security", func(w io.Writer) {
				security := space.Spec.Security
				fmt.Fprintf(w, "Developers can read logs?\t%v\n", security.EnableDeveloperLogsAccess)
			})
			fmt.Fprintln(w)

			describe.SectionWriter(w, "Build", func(w io.Writer) {
				buildpackBuild := space.Spec.BuildpackBuild
				fmt.Fprintf(w, "Builder Image:\t%q\n", buildpackBuild.BuilderImage)
				fmt.Fprintf(w, "Container Registry:\t%q\n", buildpackBuild.ContainerRegistry)
				describe.EnvVars(w, buildpackBuild.Env)
			})
			fmt.Fprintln(w)

			describe.SectionWriter(w, "Execution", func(w io.Writer) {
				execution := space.Spec.Execution
				describe.EnvVars(w, execution.Env)

				describe.SectionWriter(w, "Domains", func(w io.Writer) {
					if len(execution.Domains) == 0 {
						return
					}

					describe.TabbedWriter(w, func(w io.Writer) {
						fmt.Fprintln(w, "Name\tDefault?")
						for _, domain := range execution.Domains {
							fmt.Fprintf(w, "%s\t%t\n", domain.Domain, domain.Default)
						}
					})
				})
			})
			fmt.Fprintln(w)

			printAdditionalCommands(w, space.Name)

			return nil
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}

func printAdditionalCommands(w io.Writer, spaceName string) {
	fmt.Fprintf(w, "Use 'kf space %s' to get the current state of the space.\n", spaceName)
	fmt.Fprintf(w, "Use 'kf target -s %s' to set the default space kf works with.\n", spaceName)
	fmt.Fprintln(w, "Use 'kf configure-space' to manage the space.")
}
