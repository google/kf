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
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// NewSetEnvCommand creates a SetEnv command.
func NewSetEnvCommand(p *config.KfParams, appClient apps.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE",
		Short:   "Set an environment variable for an app",
		Example: `kf set-env myapp ENV production`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			appName := args[0]
			name := args[1]
			value := args[2]

			cmd.SilenceUsage = true

			toSet := []corev1.EnvVar{
				{Name: name, Value: value},
			}

			_, err := appClient.Transform(p.Namespace, appName, func(app *v1alpha1.App) error {
				kfapp := (*apps.KfApp)(app)
				kfapp.MergeEnvVars(toSet)

				return nil
			})

			return err
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.AppCompletion)

	return cmd
}
