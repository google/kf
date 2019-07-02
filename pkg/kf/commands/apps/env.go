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
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewEnvCommand creates a Env command.
func NewEnvCommand(p *config.KfParams, appClient apps.Client) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:     "env APP_NAME",
		Short:   "List the names and values of the environment variables for an app",
		Example: `  kf env myapp`,
		Args:    cobra.ExactArgs(1),
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			cmd.SilenceUsage = true

			app, err := appClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			kfapp := (*apps.KfApp)(app)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tVALUE")
			for _, env := range kfapp.GetEnvVars() {
				fmt.Fprintf(w, "%s\t%s\n", env.Name, env.Value)
			}
			w.Flush()

			return nil
		},
	}

	return envCmd
}
