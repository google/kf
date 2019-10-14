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
	"time"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewStartCommand creates a command capable of starting an app.
func NewStartCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:     "start APP_NAME",
		Short:   "Start a staged application",
		Example: `kf start myapp`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			cmd.SilenceUsage = true

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Stopped = false
				return nil
			}

			if _, err := client.Transform(p.Namespace, appName, mutator); err != nil {
				return fmt.Errorf("failed to start app: %s", err)
			}

			action := fmt.Sprintf("Starting app %q in space %q", appName, p.Namespace)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Namespace, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
