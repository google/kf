package apps

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewUnsetEnvCommand creates a SetEnv command.
func NewUnsetEnvCommand(p *config.KfParams, c EnvironmentClient) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:   "unset-env APP_NAME ENV_VAR_NAME",
		Short: "Unset an environment variable for an app",
		Args:  cobra.ExactArgs(2),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			name := args[1]

			cmd.SilenceUsage = true

			err := c.Unset(
				appName,
				[]string{name},
				kf.WithUnsetEnvNamespace(p.Namespace),
			)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return envCmd
}
