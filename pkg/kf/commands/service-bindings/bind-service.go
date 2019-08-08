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
	"encoding/json"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/services"
	svccatv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

// NewBindServiceCommand allows users to bind apps to service instances.
func NewBindServiceCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var (
		bindingName  string
		configAsJSON string
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

			if bindingName == "" {
				bindingName = instanceName
			}

			params, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			paramBytes, err := json.Marshal(params)
			if err != nil {
				return err
			}

			binding := &v1alpha1.AppSpecServiceBinding{
				InstanceRef: svccatv1beta1.LocalObjectReference{
					Name: instanceName,
				},
				BindingName: bindingName,
				Parameters:  paramBytes,
			}

			app, err := appsClient.Get(p.Namespace, appName)
			if err != nil {
				return err
			}

			k := apps.NewFromApp(app)
			k.BindService(binding)
			app = k.ToApp()

			_, err = appsClient.Update(p.Namespace, app)
			if err != nil {
				return err
			}

			// output.WriteBindingDetails(cmd.OutOrStdout(), binding)
			return nil
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"valid JSON object containing service-specific configuration parameters, provided in-line or in a file")

	createCmd.Flags().StringVarP(
		&bindingName,
		"binding-name",
		"b",
		"",
		"name to expose service instance to app process with (default: service instance name)")

	return createCmd
}
