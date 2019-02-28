package commands

import (
	"github.com/spf13/cobra"
)

// NewPushCommand creates a push command.
func NewPushCommand(p *KfParams) *cobra.Command {
	var push = &cobra.Command{
		Use:   "push",
		Short: "Push a new app or sync changes to an existing app",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return push
}
