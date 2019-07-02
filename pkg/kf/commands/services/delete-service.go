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

	"github.com/spf13/cobra"
)

// NewDeleteServiceCommand allows users to delete service instances.
func NewDeleteServiceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:     "delete-service SERVICE_INSTANCE",
		Aliases: []string{"ds"},
		Short:   "Delete a service instance",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]

			cmd.SilenceUsage = true

			return client.DeleteService(instanceName, services.WithDeleteServiceNamespace(p.Namespace))
		},
	}

	return deleteCmd
}
