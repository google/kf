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

package buildpacks

import (
	"fmt"
	"io"

	"github.com/google/kf/pkg/kf/buildpacks"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewStacksCommand creates a Stacks command.
func NewStacksCommand(p *config.KfParams, l buildpacks.Client) *cobra.Command {
	var buildpacksCmd = &cobra.Command{
		Use:     "stacks",
		Short:   "List stacks available in the space",
		Example: `kf stacks`,
		Args:    cobra.ExactArgs(0),
		Long: `List the stacks available in the space to applications being built
		with buildpacks.

		Stack support is determined by the buildpack builder image so they can
		change from one space to the next.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			space, err := p.GetTargetSpaceOrDefault()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Getting stacks in space: %s\n", p.Namespace)

			stacks, err := l.Stacks(space.Spec.BuildpackBuild.BuilderImage)
			if err != nil {
				cmd.SilenceUsage = !utils.ConfigError(err)
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Name")

				for _, s := range stacks {
					fmt.Fprintf(w, "%s\n", s)
				}
			})

			return nil
		},
	}

	return buildpacksCmd
}
