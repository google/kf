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

package servicebindings

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewUnbindServiceCommand allows users to unbind apps from service instances.
func NewUnbindServiceCommand(p *config.KfParams, client apps.Client) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:     "unbind-service APP_NAME SERVICE_INSTANCE",
		Aliases: []string{"us"},
		Short:   "Unbind a service instance from an app",
		Long: `Unbind removes an application's access to a service instance.

		This will delete the credential from the service broker that created the
		instance and update the VCAP_SERVICES environment variable for the
		application to remove the reference to the instance.
		`,
		Example: `kf unbind-service myapp my-instance`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			instanceName := args[1]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			if _, err := client.UnbindService(p.Namespace, appName, instanceName); err != nil {
				return err
			}

			if async.IsSynchronous() {
				fmt.Fprintf(cmd.OutOrStderr(), "Waiting for bindings to become ready on %s...\n", appName)
				if _, err := client.WaitForConditionServiceBindingsReadyTrue(context.Background(), p.Namespace, appName, 2*time.Second); err != nil {
					return fmt.Errorf("imbind failed: %s", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStderr(), "Use 'kf restart %s' to ensure your changes take effect\n", appName)

			return nil
		},
	}

	async.Add(cmd)

	return cmd
}
