package services

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/poy/service-catalog/cmd/svcat/output"

	"github.com/spf13/cobra"
)

// NewCreateServiceCommand allows users to create service instances.
func NewCreateServiceCommand(p *config.KfParams, client services.ClientInterface) *cobra.Command {
	var configAsJSON string

	createCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]",
		Aliases: []string{"cs"},
		Short:   "Create a service instance",
		Example: `
  kf create-service db-service silver mydb -c '{"ram_gb":4}'
  kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			planName := args[1]
			instanceName := args[2]

			cmd.SilenceUsage = true

			params, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			instance, err := client.CreateService(
				instanceName,
				serviceName,
				planName,
				services.WithCreateServiceNamespace(p.Namespace),
				services.WithCreateServiceParams(params))
			if err != nil {
				return err
			}

			output.WriteInstanceDetails(p.Output, instance)
			return nil
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"Valid JSON object containing service-specific configuration parameters, provided in-line or in a file.")

	createCmd.SetOutput(p.Output)
	return createCmd
}
