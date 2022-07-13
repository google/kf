package exporttok8s

import (
	"fmt"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

func NewExportToK8s(cfg *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Use:     "export-to-k8s",
		Short:   "export yaml files for the app",
		Example: `kf export-to-k8s`,
		Args:    cobra.ExactArgs(0),
		Long: `
		The export-to-k8s command allows operators to export the Tekton Pipeline, PipelineRun
		and App deployment files.

		Users can edit and execute the Tekton yaml files. The pipeline would then export an
		App image URL. Users need to replace the image in the deployment file with the
		exported image URL, and execute the deployment file to deploy their App.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "This is the new command export-to-k8s!!")
			return nil
		},
	}
	return cmd
}
