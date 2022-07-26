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

	"github.com/google/kf/v2/pkg/kf/buildpacks"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewBuildpacksCommand creates a Buildpacks command.
func NewBuildpacksCommand(p *config.KfParams, l buildpacks.Client) *cobra.Command {
	var buildpacksCmd = &cobra.Command{
		Use:     "buildpacks",
		Short:   "List buildpacks in the targeted Space.",
		Args:    cobra.ExactArgs(0),
		Example: "kf buildpacks",
		Long: `
		List the buildpacks available in the Space to Apps being built with
		buildpacks.

		The buildpacks available to an App depend on the Stack it uses.
		To ensure reproducibility in Builds, Apps should explicitly declare the
		Stack they use.
		`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			space, err := p.GetTargetSpace(ctx)
			if err != nil {
				return err
			}

			logging.FromContext(ctx).Infof("Getting buildpacks in Space: %s", p.Space)

			if p.FeatureFlags(ctx).AppDevExperienceBuilds().IsDisabled() {
				describe.SectionWriter(cmd.OutOrStdout(), "Buildpacks for V2 Stacks", func(w io.Writer) {
					fmt.Fprintln(w, "Name\tPosition\tURL")

					for i, bp := range space.Status.BuildConfig.BuildpacksV2 {
						fmt.Fprintf(w, "%s\t%d\t%s\n", bp.Name, i, bp.URL)
					}
				})
			}

			for _, stack := range space.Status.BuildConfig.StacksV3 {
				describe.SectionWriter(cmd.OutOrStdout(), "V3 Stack: "+stack.Name, func(w io.Writer) {
					bps, err := l.List(stack.BuildImage)
					if err != nil {
						fmt.Fprintf(w, "error fetching stacks for image %s: %s\n", stack.BuildImage, err)
						return
					}

					describe.TabbedWriter(w, func(w io.Writer) {
						fmt.Fprintln(w, "Name\tPosition\tVersion\tLatest")

						for i, bp := range bps {
							fmt.Fprintf(w, "%s\t%d\t%s\t%v\n", bp.ID, i, bp.Version, bp.Latest)
						}
					})
				})
			}

			return nil
		},
	}

	return buildpacksCmd
}
