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
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/services"
	"github.com/spf13/cobra"
)

// NewGetServiceCommand allows users to get a service instance.
func NewGetServiceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	serviceCommand := &cobra.Command{
		Use:     "service SERVICE_INSTANCE",
		Short:   "Show service instance info",
		Example: `kf service my-service`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			instanceName := args[0]

			cmd.SilenceUsage = true

			instance, err := client.GetService(instanceName, services.WithGetServiceNamespace(p.Namespace))
			if err != nil {
				return err
			}

			describe.ServiceInstance(cmd.OutOrStdout(), instance)

			return nil
		},
	}

	return serviceCommand
}
