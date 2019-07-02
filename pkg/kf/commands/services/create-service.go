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

package services

import (
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/services"
	"github.com/poy/service-catalog/cmd/svcat/output"

	"github.com/spf13/cobra"
)

// NewCreateServiceCommand allows users to create service instances.
func NewCreateServiceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	var configAsJSON string

	createCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]",
		Aliases: []string{"cs"},
		Short:   "Create a service instance",
		Example: `
  kf create-service db-service silver mydb -c '{"ram_gb":4}'
  kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			planName := args[1]
			instanceName := args[2]

			cmd.SilenceUsage = true

			params, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			instance, err := client.CreateService(
				instanceName,
				serviceName,
				planName,
				services.WithCreateServiceNamespace(p.Namespace),
				services.WithCreateServiceParams(params))
			if err != nil {
				return err
			}

			output.WriteInstanceDetails(cmd.OutOrStdout(), instance)
			return nil
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"Valid JSON object containing service-specific configuration parameters, provided in-line or in a file.")

	return createCmd
}
