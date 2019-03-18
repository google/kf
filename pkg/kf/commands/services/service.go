package services

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
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
				fmt.Fprintf(p.Output, "service %s not found", instanceName)
			} else {
				output.WriteInstance(p.Output, "table", *instance)
			}

			return nil
		},
	}

	return serviceCommand
}
