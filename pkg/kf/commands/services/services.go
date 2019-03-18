package services

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
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

			output.WriteInstanceList(p.Output, "table", instances)

			return nil
		},
	}

	return servicesCommand
}
