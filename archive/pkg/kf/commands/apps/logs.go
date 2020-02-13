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

package apps

import (
	"context"
	"fmt"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/logs"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates a Logs command.
func NewLogsCommand(p *config.KfParams, tailer logs.Tailer) *cobra.Command {
	var (
		numberLines int
		recent      bool
	)
	cmd := &cobra.Command{
		Use:   "logs APP_NAME",
		Short: "Tail or show logs for an app",
		Example: `
		# Follow/tail the log stream
		kf logs myapp

		# Follow/tail the log stream with 20 lines of context
		kf logs myapp -n 20

		# Get recent logs from the app
		kf logs myapp --recent

		# Get the most recent 200 lines of logs from the app
		kf logs myapp --recent -n 200
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			shouldFollow := !recent

			appName := args[0]
			if err := tailer.Tail(
				context.Background(),
				appName,
				cmd.OutOrStdout(),
				logs.WithTailNamespace(p.Namespace),
				logs.WithTailNumberLines(numberLines),
				logs.WithTailFollow(shouldFollow),
			); err != nil {
				cmd.SilenceUsage = !utils.ConfigError(err)
				return fmt.Errorf("failed to tail logs: %s", err)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(
		&numberLines,
		"number",
		"n",
		10,
		"Show the last N lines of logs.",
	)

	cmd.Flags().BoolVarP(
		&recent,
		"recent",
		"",
		false,
		"Dump recent logs instead of tailing.",
	)

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
