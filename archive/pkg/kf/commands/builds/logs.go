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
	"context"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/sources"
	"github.com/spf13/cobra"
)

// NewBuildLogsCommand allows users to list spaces.
func NewBuildLogsCommand(p *config.KfParams, client sources.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "build-logs BUILD_NAME",
		Short:   "Get the logs of the given build",
		Example: "kf build-logs build-12345",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			buildName := args[0]

			return client.Tail(context.Background(), p.Namespace, buildName, cmd.OutOrStdout())
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.SourceCompletion)

	return cmd
}
