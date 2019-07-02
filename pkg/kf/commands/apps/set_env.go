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
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// NewSetEnvCommand creates a SetEnv command.
func NewSetEnvCommand(p *config.KfParams, appClient apps.Client) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:     "set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE",
		Short:   "Set an environment variable for an app",
		Example: `  kf set-env myapp FOO bar`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			name := args[1]
			value := args[2]

			cmd.SilenceUsage = true

			toSet := []corev1.EnvVar{
				{Name: name, Value: value},
			}

			return appClient.Transform(p.Namespace, appName, func(app *serving.Service) error {
				kfapp := (*apps.KfApp)(app)
				kfapp.MergeEnvVars(toSet)

				return nil
			})
		},
	}

	return envCmd
}
