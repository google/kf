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

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/services"
	"github.com/spf13/cobra"
)

// NewBindServiceCommand allows users to bind apps to service instances.
func NewBindServiceCommand(p *config.KfParams, client apps.Client) *cobra.Command {
	var (
		bindingName  string
		configAsJSON string
		async        utils.AsyncFlags
	)

	createCmd := &cobra.Command{
		Use:     "bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME]",
		Aliases: []string{"bs"},
		Short:   "Bind a service instance to an app",
		Example: `  kf bind-service myapp mydb -c '{"permissions":"read-only"}'`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			instanceName := args[1]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			parameters, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			binding := &v1alpha1.AppSpecServiceBinding{
				Instance:    instanceName,
				Parameters:  parameters,
				BindingName: bindingName,
			}

			if _, err := client.BindService(p.Namespace, appName, binding); err != nil {
				return err
			}

			if async.IsSynchronous() {
				fmt.Fprintf(cmd.OutOrStderr(), "Waiting for bindings to become ready on %s...\n", appName)
				if _, err := client.WaitForConditionServiceBindingsReadyTrue(context.Background(), p.Namespace, appName, 2*time.Second); err != nil {
					return fmt.Errorf("bind failed: %s", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStderr(), "Use 'kf restart %s' to ensure your changes take effect\n", appName)

			return nil
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"JSON object containing service-specific configuration parameters, provided in-line or in a file")

	createCmd.Flags().StringVarP(
		&bindingName,
		"binding-name",
		"b",
		"",
		"Name to expose service instance to app process with (default: service instance name)")

	async.Add(createCmd)

	return createCmd
}
