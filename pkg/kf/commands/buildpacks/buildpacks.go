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

// NewBuildpacksCommand creates a Buildpacks command.
func NewBuildpacksCommand(p *config.KfParams, l buildpacks.Client) *cobra.Command {
	var buildpacksCmd = &cobra.Command{
		Use:     "buildpacks",
		Short:   "List buildpacks in current builder",
		Args:    cobra.ExactArgs(0),
		Example: "kf buildpacks",
		Long: `List the buildpacks available in the space to applications being built
		with buildpacks.

		Buildpack support is determined by the buildpack builder image and can
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

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Getting buildpacks in space: %s\n", p.Namespace); err != nil {
				return err
			}

			bps, err := l.List(space.Spec.BuildpackBuild.BuilderImage)
			if err != nil {
				cmd.SilenceUsage = !utils.ConfigError(err)
				return err
			}

			if err := describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) error {
				if _, err := fmt.Fprintln(w, "Name\tPosition\tVersion\tLatest"); err != nil {
					return err
				}

				for i, bp := range bps {
					if _, err := fmt.Fprintf(w, "%s\t%d\t%s\t%v\n", bp.ID, i, bp.Version, bp.Latest); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return buildpacksCmd
}
