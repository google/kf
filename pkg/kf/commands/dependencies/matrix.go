// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dependencies

import (
	"fmt"

	"text/tabwriter"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// newMatrixCommand returns a command that lists dependencies
// in a markdown matrix
func newMatrixCommand(dependencies []dependency) *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		Use:          "matrix",
		Short:        "Get Kf dependencies and links in a publishable format",
		Long:         documentationOnly,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
			defer out.Flush()

			fmt.Fprintln(out, "| Dependency\t| Version\t|")
			fmt.Fprintln(out, "| ---\t| ---\t|")

			for _, dep := range dependencies {

				version, _, err := dep.ResolveAll()
				if err != nil {
					return err
				}

				fmt.Fprintf(out, "| [%s](%s)\t| `%s`\t|",
					dep.Name,
					dep.InfoURL,
					version,
				)
				fmt.Fprintln(out)
			}
			return nil
		},
	}

	return cmd
}
