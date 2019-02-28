package commands

import (
	"github.com/spf13/cobra"
)

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "kf is like cf for Knative",
		Long: `kf is like cf for Knative

	Supported sub-commands are:

	  kf push
	  kf delete <app>
	  kf apps

	You can get more info by adding the --help flag to any sub-command.
	  `,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	// Model new commands after:
	// https://github.com/knative/client/blob/master/pkg/kn/commands/service_list.go
	// to take an idiomatic k8s like approach.
	rootCmd.AddCommand(NewPushCommand())

	return rootCmd
}
