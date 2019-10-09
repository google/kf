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
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/services"
	"github.com/spf13/cobra"
)

// NewDeleteServiceCommand allows users to delete service instances.
func NewDeleteServiceCommand(p *config.KfParams, client services.Client) *cobra.Command {
	var async utils.AsyncFlags

	deleteCmd := &cobra.Command{
		Use:     "delete-service SERVICE_INSTANCE",
		Aliases: []string{"ds"},
		Short:   "Delete a service instance",
		Example: "kf delete-service my-service",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			if err := client.Delete(p.Namespace, instanceName); err != nil {
				return err
			}

			action := fmt.Sprintf("Deleting service instance %q in space %q", instanceName, p.Namespace)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForDeletion(context.Background(), p.Namespace, instanceName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(deleteCmd)

	return deleteCmd
}
