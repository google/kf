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
	"io"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/kf/marketplace"
	"github.com/spf13/cobra"
)

// NewMarketplaceCommand allows users to get a service instance.
func NewMarketplaceCommand(p *config.KfParams, marketplaceClient marketplace.ClientInterface) *cobra.Command {
	var serviceName string

	marketplaceCommand := &cobra.Command{
		Use:     "marketplace [-s SERVICE]",
		Aliases: []string{"m"},
		Short:   "List service classes available in the cluster.",
		Example: `
		# Show service classes available in the cluster
		kf marketplace

		# Show the plans available to a particular service class
		kf marketplace -s google-storage
		`,
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			catalog, err := marketplaceClient.Marketplace(cmd.Context(), p.Space)
			if err != nil {
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				if serviceName == "" {
					fmt.Fprintf(w, "Listing services that can be used in Space %q, use the --service flag to list the plans for a service\n", p.Space)
					fmt.Fprintln(w)

					fmt.Fprintln(w, "Broker\tName\tNamespace\tDescription")
					catalog.WalkServiceOfferings(func(lineage marketplace.OfferingLineage) {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
							lineage.Broker.GetName(),
							lineage.ServiceOffering.DisplayName,
							lineage.Broker.GetNamespace(),
							lineage.ServiceOffering.Description,
						)
					})
				} else {
					fmt.Fprintln(w, "Broker\tName\tFree\tDescription")
					catalog.WalkServicePlans(func(lineage marketplace.PlanLineage) {
						if lineage.ServiceOffering.DisplayName != serviceName {
							return
						}

						fmt.Fprintf(w, "%s\t%s\t%t\t%s\n",
							lineage.Broker.GetName(),
							lineage.ServicePlan.DisplayName,
							lineage.ServicePlan.Free,
							lineage.ServicePlan.Description,
						)
					})
				}
			})

			return nil
		},
	}

	// TODO: there should be a verbose option here to dump full info.
	marketplaceCommand.Flags().StringVarP(
		&serviceName,
		"service",
		"s",
		"",
		"List plans for the service class.")

	return marketplaceCommand
}
