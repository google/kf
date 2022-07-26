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

package builds

import (
	"github.com/google/kf/v2/pkg/kf/builds"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewBuildLogsCommand allows users to list Spaces.
func NewBuildLogsCommand(p *config.KfParams, client builds.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "build-logs BUILD_NAME",
		Short:             "Get the logs of the given Build.",
		Example:           "kf build-logs build-12345",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.BuildCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			buildName := args[0]

			return client.Tail(ctx, p.Space, buildName, cmd.OutOrStdout())
		},
	}

	return cmd
}
