package servicebindings

import (
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/spf13/cobra"
)

// NewVcapServicesCommand allows users to bind apps to service instances.
func NewVcapServicesCommand(p *config.KfParams, client servicebindings.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcap-services APP_NAME",
		Short: "Print the VCAP_SERVICES environment variable for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]

			cmd.SilenceUsage = true

			output, err := client.GetVcapServices(appName,
				servicebindings.WithGetVcapServicesNamespace(p.Namespace))
			if err != nil {
				return err
			}

			out, err := json.Marshal(output)
			if err != nil {
				return err
			}

			fmt.Fprintln(p.Output, string(out))

			return nil
		},
	}

	return cmd
}
