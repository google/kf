package apps

import (
	"path/filepath"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	kfi "github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/spf13/cobra"
)

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName string, opts ...kf.PushOption) error
}

// NewPushCommand creates a push command.
func NewPushCommand(p *config.KfParams, pusher Pusher) *cobra.Command {
	var (
		containerRegistry string
		dockerImage       string
		serviceAccount    string
		path              string
	)

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push a new app or sync changes to an existing app",
		Args:  cobra.ExactArgs(1),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			if path != "" {
				var err error
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
			}

			err := pusher.Push(
				appName,
				kf.WithPushNamespace(p.Namespace),
				kf.WithPushContainerRegistry(containerRegistry),
				kf.WithPushDockerImage(dockerImage),
				kf.WithPushServiceAccount(serviceAccount),
				kf.WithPushPath(path),
			)
			cmd.SilenceUsage = !kfi.ConfigError(err)

			return err
		},
	}

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push containers. Either docker-image or container-registry must be set (but not both).",
	)

	pushCmd.Flags().StringVar(
		&dockerImage,
		"docker-image",
		"",
		"The docker image to push. Either docker-image or container-registry must be set (but not both).",
	)

	pushCmd.Flags().StringVar(
		&serviceAccount,
		"service-account",
		"",
		"The service account to enable access to the container registry (REQUIRED)",
	)

	pushCmd.Flags().StringVar(
		&path,
		"path",
		"",
		"The path the source code lives. Defaults to current directory.",
	)

	return pushCmd
}
