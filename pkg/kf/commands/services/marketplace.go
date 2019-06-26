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
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/services"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"

	"github.com/spf13/cobra"
)

// NewMarketplaceCommand allows users to get a service instance.
func NewMarketplaceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	var serviceName string

	marketplaceCommand := &cobra.Command{
		Use:     "marketplace [-s SERVICE]",
		Aliases: []string{"m"},
		Short:   "List available offerings in the marketplace",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			marketplace, err := client.Marketplace(services.WithMarketplaceNamespace(p.Namespace))
			if err != nil {
				return err
			}

			// We use a custom tabwriter rather than svcat outputs because the
			// headings on there don't make sense for our target audience.
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 2, ' ', 0)

			if serviceName == "" {
				fmt.Fprintf(w, "%d services can be used in namespace %q, use the --service flag to list the plans for a service\n", len(marketplace.Services), p.Namespace)
				fmt.Fprintln(w)

				fmt.Fprintln(w, "BROKER\tNAME\tNAMESPACE\tSTATUS\tDESCRIPTION")
				for _, s := range marketplace.Services {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.100s\n", s.GetServiceBrokerName(), s.GetExternalName(), s.GetNamespace(), s.GetStatusText(), s.GetDescription())
				}
			} else {
				// If the user wants to show service plans by a service name, then
				// we MUST convert that into a "ClassID" which corresponds with
				// service GUID in the OSB spec by translating it using the list of
				// classes (services) first.
				serviceGUID := serviceName
				for _, service := range marketplace.Services {
					if service.GetExternalName() == serviceName {
						serviceGUID = service.GetName()
						break
					}
				}

				var filteredPlans []servicecatalog.Plan
				for _, plan := range marketplace.Plans {
					if plan.GetClassID() == serviceGUID {
						filteredPlans = append(filteredPlans, plan)
					}
				}

				fmt.Fprintln(w, "NAME\tFREE\tSTATUS\tDESCRIPTION")
				for _, p := range filteredPlans {
					fmt.Fprintf(w, "%s\t%t\t%s\t%.100s\n", p.GetExternalName(), p.GetFree(), p.GetShortStatus(), p.GetDescription())
				}
			}

			w.Flush()

			return nil
		},
	}

	// TODO there should be a verbose option here to dump full info.
	marketplaceCommand.Flags().StringVarP(
		&serviceName,
		"service",
		"s",
		"",
		"Show plan details for a particular service offering.")

	return marketplaceCommand
}
