package services

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"

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
