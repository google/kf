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
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	cluster "github.com/google/kf/v2/pkg/kf/service-brokers/cluster"
	namespaced "github.com/google/kf/v2/pkg/kf/service-brokers/namespaced"
	"github.com/spf13/cobra"
)

// NewDeleteServiceBrokerCommand deletes a service broker (either cluster or namespaced) from the service catalog.
func NewDeleteServiceBrokerCommand(
	p *config.KfParams,
	clusterClient cluster.Client,
	namespacedClient namespaced.Client,
) *cobra.Command {
	var (
		spaceScoped bool
		async       utils.AsyncFlags
	)
	deleteCmd := &cobra.Command{
		Use:          "delete-service-broker NAME",
		Aliases:      []string{"dsb"},
		Short:        "Remove a service broker from the marketplace.",
		Example:      `  kf delete-service-broker mybroker`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceBrokerName := args[0]

			if spaceScoped {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
			}

			var callback func() error
			switch {
			case spaceScoped:
				if err := namespacedClient.Delete(cmd.Context(), p.Space, serviceBrokerName); err != nil {
					return err
				}

				callback = func() (err error) {
					_, err = namespacedClient.WaitForDeletion(cmd.Context(), p.Space, serviceBrokerName, 1*time.Second)
					return
				}

			default:
				if err := clusterClient.Delete(cmd.Context(), serviceBrokerName); err != nil {
					return err
				}

				callback = func() (err error) {
					_, err = clusterClient.WaitForDeletion(cmd.Context(), serviceBrokerName, 1*time.Second)
					return
				}
			}

			// Set up messages for the user
			var action string
			if spaceScoped {
				action = fmt.Sprintf("Deleting service broker %q in Space %q", serviceBrokerName, p.Space)
			} else {
				action = fmt.Sprintf("Deleting cluster service broker %q", serviceBrokerName)
			}

			return async.AwaitAndLog(cmd.OutOrStdout(), action, callback)
		},
	}

	async.Add(deleteCmd)

	deleteCmd.Flags().BoolVar(
		&spaceScoped,
		"space-scoped",
		false,
		"Set to delete a space scoped service broker.")

	return deleteCmd
}
