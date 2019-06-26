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

	"github.com/google/kf/pkg/kf/commands/config"
	kfi "github.com/google/kf/pkg/kf/internal/kf"
	"github.com/google/kf/pkg/kf/logs"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates a Logs command.
func NewLogsCommand(p *config.KfParams, tailer logs.Tailer) *cobra.Command {
	var (
		numberLines int
		follow      bool
	)
	c := &cobra.Command{
		Use:   "logs APP_NAME",
		Short: "View or follow logs for an app",
		Example: `
  kf logs myapp
  kf logs myapp -n 20
  kf logs myapp -f
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			if err := tailer.Tail(
				context.Background(),
				appName,
				cmd.OutOrStdout(),
				logs.WithTailNamespace(p.Namespace),
				logs.WithTailNumberLines(numberLines),
				logs.WithTailFollow(follow),
			); err != nil {
				cmd.SilenceUsage = !kfi.ConfigError(err)
				return fmt.Errorf("failed to tail logs: %s", err)
			}

			return nil
		},
	}

	c.Flags().IntVarP(
		&numberLines,
		"number",
		"n",
		10,
		"The number of lines from the end of the logs to show.",
	)
	c.Flags().BoolVarP(
		&follow,
		"follow",
		"f",
		false,
		"Follow the log stream of the app.",
	)

	return c
}
