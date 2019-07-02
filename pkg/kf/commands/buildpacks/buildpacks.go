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
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/buildpacks"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/kf"
	"github.com/spf13/cobra"
)

// NewBuildpacksCommand creates a Buildpacks command.
func NewBuildpacksCommand(p *config.KfParams, l buildpacks.Client) *cobra.Command {
	var buildpacksCmd = &cobra.Command{
		Use:   "buildpacks",
		Short: "List buildpacks in current builder.",
		Args:  cobra.ExactArgs(0),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if p.Namespace != "default" && p.Namespace != "" {
				fmt.Fprintf(cmd.OutOrStderr(), "NOTE: Buildpacks are global and are available to all spaces.")
			}
			bps, err := l.List()
			if err != nil {
				cmd.SilenceUsage = !kf.ConfigError(err)
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tPOSITION\tVERSION\tLATEST")
			for i, bp := range bps {
				fmt.Fprintf(w, "%s\t%d\t%s\t%v\n", bp.ID, i, bp.Version, bp.Latest)
			}
			w.Flush()

			return nil
		},
	}

	return buildpacksCmd
}
