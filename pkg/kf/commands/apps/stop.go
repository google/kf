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
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
)

// NewStopCommand creates a command capable of stopping an app.
func NewStopCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stop APP_NAME",
		Short:   "Stop a running application",
		Example: `kf stop myapp`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]

			cmd.SilenceUsage = true

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Stopped = true
				return nil
			}

			if _, err := client.Transform(p.Namespace, appName, mutator); err != nil {
				return fmt.Errorf("failed to stop app: %s", err)
			}

			return nil
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
