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
	"context"
	"fmt"
	"io"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewStacksCommand creates a Stacks command.
func NewStacksCommand(p *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stacks",
		Short:   "List stacks in the targeted Space.",
		Example: `kf stacks`,
		Args:    cobra.ExactArgs(0),
		Long: `
		Stacks contain information about how to build and run an App.
		Each stack contains:

		*  A unique name to identify it.
		*  A build image, the image used to build the App, this usually contains
			 things like compilers, libraries, sources and build frameworks.
		*  A run image, the image App instances will run within. These images
			 are usually lightweight and contain just enough to run an App.
		*  A list of applicable buildpacks available via the bulidpacks command.
		`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			space, err := p.GetTargetSpace(context.Background())
			if err != nil {
				return err
			}

			logging.FromContext(ctx).Infof("Getting stacks in Space: %s", p.Space)

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Version\tName\tBuild Image\tRun Image\tDescription")

				if p.FeatureFlags(ctx).AppDevExperienceBuilds().IsDisabled() {
					for _, s := range space.Status.BuildConfig.StacksV2 {
						fmt.Fprintf(w, "V2\t%s\t%s\t%s\t%s\n", s.Name, s.Image, s.Image, s.Description)
					}
				}

				for _, s := range space.Status.BuildConfig.StacksV3 {
					fmt.Fprintf(w, "V3\t%s\t%s\t%s\t%s\n", s.Name, s.BuildImage, s.RunImage, s.Description)
				}
			})

			return nil
		},
	}

	return cmd
}
