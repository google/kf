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

	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewRestageCommand creates a command capable of restaging an app.
func NewRestageCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:   "restage APP_NAME",
		Short: "Rebuild and redeploy an App without downtime.",
		Long: `
		Restaging an App will re-run the latest Build to produce a new
		container image, and if successful will replace each running instance
		with the new image.

		Instances are replaced one at a time, always ensuring that the desired
		number of instances are healthy. This property is upheld by running one
		additional instance of the App and swapping it out for an old instance.

		The operation completes once all instances have been replaced.
		`,
		Example:           `kf restage myapp`,
		Args:              cobra.ExactArgs(1),
		Aliases:           []string{"rg"},
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			app, err := client.Restage(cmd.Context(), p.Space, appName)
			if err != nil {
				return fmt.Errorf("failed to restage App: %s", err)
			}

			if async.IsSynchronous() {
				if err := client.DeployLogsForApp(cmd.Context(), cmd.OutOrStdout(), app); err != nil {
					return fmt.Errorf("failed to restage App: %s", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%q successfully restaged\n", appName)
			}

			return nil
		},
	}

	async.Add(cmd)

	return cmd
}
