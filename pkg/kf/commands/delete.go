package commands

import (
	"log"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/spf13/cobra"
)

// Deleter deletes deployed apps.
type Deleter interface {
	// Delete deletes deployed apps.
	Delete(appName string, opts ...kf.DeleteOption) error
}

// NewDeleteCommand creates a delete command.
func NewDeleteCommand(p *KfParams, d Deleter) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing app",
		Args:  cobra.ExactArgs(1),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			if err := d.Delete(appName, kf.WithDeleteNamespace(p.Namespace)); err != nil {
				return err
			}

			log.Printf("app %q has been successfully deleted\n", appName)
			return nil
		},
	}

	return deleteCmd
}
