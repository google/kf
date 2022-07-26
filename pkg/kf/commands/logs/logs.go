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

package logs

import (
	"context"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/logs"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates a Logs command.
func NewLogsCommand(p *config.KfParams, tailer logs.Tailer) *cobra.Command {
	var (
		numberLines int
		recent      bool
		task        bool
	)
	cmd := &cobra.Command{
		Use:   "logs APP_NAME",
		Short: "Show logs for an App.",
		Long: `Logs are streamed from the Kubernetes log endpoint for each running
		App instance.

		If App instances change or the connection to Kubernetes times out the
		log stream may show duplicate logs.

		Logs are retained for App instances as space permits on the cluster,
		but will be deleted if space is low or past their retention date.
		Cloud Logging is a more reliable mechanism to access historical logs.

		If you need logs for a particular instance use the <code>kubectl</code> CLI.
		`,
		Example: `
		# Follow/tail the log stream
		kf logs myapp

		# Follow/tail the log stream with 20 lines of context
		kf logs myapp -n 20

		# Get recent logs from the App
		kf logs myapp --recent

		# Get the most recent 200 lines of logs from the App
		kf logs myapp --recent -n 200

		# Get the logs of Tasks running from the App
		kf logs myapp --task
		`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			shouldFollow := !recent

			componentName := "app-server"
			containerName := v1alpha1.DefaultUserContainerName
			labels := make(map[string]string)
			if task == true {
				componentName = "task"
				containerName = fmt.Sprintf("step-%s", v1alpha1.DefaultUserContainerName)
				labels["tekton.dev/pipelineTask"] = v1alpha1.DefaultUserContainerName
			}

			appName := args[0]
			if err := tailer.Tail(
				context.Background(),
				appName,
				cmd.OutOrStdout(),
				logs.WithTailSpace(p.Space),
				logs.WithTailNumberLines(numberLines),
				logs.WithTailFollow(shouldFollow),
				logs.WithTailComponentName(componentName),
				logs.WithTailContainerName(containerName),
				logs.WithTailLabels(labels),
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

	cmd.Flags().BoolVarP(
		&task,
		"task",
		"",
		false,
		"Tail Task logs instead of App.",
	)

	return cmd
}
