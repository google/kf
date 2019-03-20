package services

import (
	"fmt"
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
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
			w := tabwriter.NewWriter(p.Output, 8, 4, 2, ' ', 0)

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
