package apps

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewSetEnvCommand creates a SetEnv command.
func NewSetEnvCommand(p *config.KfParams, c EnvironmentClient) *cobra.Command {
	var envCmd = &cobra.Command{
		Use:   "set-env APP_NAME [NAME=VALUE]",
		Short: "Set an environment variable for an app",
		Args:  cobra.ExactArgs(3),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			name := args[1]
			value := args[2]

			cmd.SilenceUsage = true

			err := c.Set(
				appName,
				map[string]string{name: value},
				kf.WithSetEnvNamespace(p.Namespace),
			)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return envCmd
}
