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
	"fmt"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/services"
	"github.com/poy/service-catalog/cmd/svcat/output"

	"github.com/spf13/cobra"
)

// NewGetServiceCommand allows users to get a service instance.
func NewGetServiceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	serviceCommand := &cobra.Command{
		Use:   "service SERVICE_INSTANCE",
		Short: "Show service instance info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]

			cmd.SilenceUsage = true

			instance, err := client.GetService(instanceName, services.WithGetServiceNamespace(p.Namespace))
			if err != nil {
				return err
			}

			if instance == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "service %s not found", instanceName)
			} else {
				output.WriteInstance(cmd.OutOrStdout(), "table", *instance)
			}

			return nil
		},
	}

	return serviceCommand
}
