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

package servicebrokers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/kf/commands/config"
	installutil "github.com/google/kf/pkg/kf/commands/install/util"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/service-brokers/cluster"
	"github.com/google/kf/pkg/kf/service-brokers/namespaced"
	"github.com/spf13/cobra"
)

// NewDeleteServiceBrokerCommand deletes a service broker (either cluster or namespaced) from the service catalog.
func NewDeleteServiceBrokerCommand(p *config.KfParams, clusterClient cluster.Client, namespacedClient namespaced.Client) *cobra.Command {
	var (
		spaceScoped bool
		force       bool
		async       utils.AsyncFlags
	)
	deleteCmd := &cobra.Command{
		Use:     "delete-service-broker BROKER_NAME",
		Aliases: []string{"dsb"},
		Short:   "Remove a service broker from service catalog",
		Example: `  kf delete-service-broker mybroker`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceBrokerName := args[0]

			cmd.SilenceUsage = true

			if !force {
				shouldDelete, err := installutil.SelectYesNo(context.Background(), fmt.Sprintf("Really delete service-broker %s?", serviceBrokerName))
				if err != nil || shouldDelete == false {
					fmt.Fprintln(cmd.OutOrStdout(), "Skipping deletion, use --force to delete without validation")
					return err
				}
			}

			if spaceScoped {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}

				if err := namespacedClient.Delete(p.Namespace, serviceBrokerName); err != nil {
					return err
				}

				action := fmt.Sprintf("Deleting service broker %q in space %q", serviceBrokerName, p.Namespace)
				return async.AwaitAndLog(cmd.OutOrStdout(), action, func() (err error) {
					_, err = namespacedClient.WaitForDeletion(context.Background(), p.Namespace, serviceBrokerName, 1*time.Second)
					return
				})
			}

			if err := clusterClient.Delete(serviceBrokerName); err != nil {
				return err
			}

			action := fmt.Sprintf("Deleting cluster service broker %q", serviceBrokerName)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() (err error) {
				_, err = clusterClient.WaitForDeletion(context.Background(), serviceBrokerName, 1*time.Second)
				return
			})
		},
	}

	async.Add(deleteCmd)

	deleteCmd.Flags().BoolVar(
		&spaceScoped,
		"space-scoped",
		false,
		"Set to delete a space scoped service broker.")

	deleteCmd.Flags().BoolVar(
		&force,
		"force",
		false,
		"Set to force deletion without a confirmation prompt.")

	return deleteCmd
}
