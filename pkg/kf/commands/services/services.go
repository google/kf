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

// NewListServicesCommand allows users to list service instances.
func NewListServicesCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	servicesCommand := &cobra.Command{
		Use:     "services",
		Aliases: []string{"s"},
		Short:   "List all service instances in the target namespace",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {

			cmd.SilenceUsage = true

			instances, err := client.ListServices(services.WithListServicesNamespace(p.Namespace))
			if err != nil {
				return err
			}

			output.WriteInstanceList(cmd.OutOrStdout(), "table", instances)

			return nil
		},
	}

	return servicesCommand
}
