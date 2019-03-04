package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeleteCommand creates a delete command.
func NewDeleteCommand(p *KfParams) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing app",
		Args:  cobra.ExactArgs(1),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			cmd.SilenceUsage = true
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			propPolicy := k8smeta.DeletePropagationForeground
			if err := client.Services(p.Namespace).Delete(appName, &k8smeta.DeleteOptions{
				PropagationPolicy: &propPolicy,
			}); err != nil {
				return err
			}

			fmt.Fprintf(p.Output, "app %q has been successfully deleted\n", appName)
			return nil
		},
	}

	return deleteCmd
}
