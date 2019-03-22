package servicebindings

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"

	"github.com/spf13/cobra"
)

// NewUnbindServiceCommand allows users to bind apps to service instances.
func NewUnbindServiceCommand(p *config.KfParams, client servicebindings.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unbind-service APP_NAME SERVICE_INSTANCE",
		Aliases: []string{"us"},
		Short:   "Unbind a service instance from an app",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			instanceName := args[1]

			cmd.SilenceUsage = true

			err := client.Delete(
				instanceName,
				appName,
				servicebindings.WithDeleteNamespace(p.Namespace))
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
