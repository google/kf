package commands

import (
	"errors"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/spf13/cobra"
)

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName string, opts ...kf.PushOption) error
}

// NewPushCommand creates a push command.
func NewPushCommand(p *KfParams, pusher Pusher) *cobra.Command {
	var (
		containerRegistry string
		serviceAccount    string
	)

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push a new app or sync changes to an existing app",
		Args:  cobra.ExactArgs(1),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra ensures we are only called with a single argument.
			appName := args[0]
			if containerRegistry == "" {
				return errors.New("container registry is not set")
			}
			if serviceAccount == "" {
				return errors.New("service account is not set")
			}

			cmd.SilenceUsage = true
			return pusher.Push(
				appName,
				kf.WithPushNamespace(p.Namespace),
				kf.WithPushContainerRegistry(containerRegistry),
				kf.WithPushServiceAccount(serviceAccount),
			)
		},
	}

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push containers (REQUIRED)",
	)

	pushCmd.Flags().StringVar(
		&serviceAccount,
		"service-account",
		"",
		"The service account to enable access to the container registry (REQUIRED)",
	)

	return pushCmd
}
