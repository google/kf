package buildpacks

import (
	"fmt"
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/spf13/cobra"
)

// BuildpackLister lists the buildpacks available.
type BuildpackLister interface {
	// List lists the buildpacks available.
	List(opts ...buildpacks.BuildpackListOption) ([]string, error)
}

// NewUploadBuildpacks creates a UploadBuildpacks command.
func NewBuildpacks(p *config.KfParams, l BuildpackLister) *cobra.Command {
	var buildpacksCmd = &cobra.Command{
		Use:   "buildpacks",
		Short: "List buildpacks in current builder.",
		Args:  cobra.ExactArgs(0),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			bps, err := l.List(
				buildpacks.WithBuildpackListNamespace(p.Namespace),
			)
			if err != nil {
				cmd.SilenceUsage = !kf.ConfigError(err)
				return err
			}

			w := tabwriter.NewWriter(p.Output, 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tPOSITION\tENABLED")
			for i, bp := range bps {
				fmt.Fprintf(w, "%s\t%d\t%v\n", bp, i, true)
			}
			w.Flush()

			return nil
		},
	}

	return buildpacksCmd
}
