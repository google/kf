package commands

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/spf13/cobra"
)

// AppLister lists deployed applications.
type AppLister interface {
	// List the deployed applications in a namespace.
	List(namespace string) ([]kf.App, error)
}

// NewAppsCommand creates a apps command.
func NewAppsCommand(p *KfParams, l AppLister) *cobra.Command {
	var apps = &cobra.Command{
		Use:   "apps",
		Short: "List pushed apps.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			apps, err := l.List(p.Namespace)
			if err != nil {
				return err
			}

			for _, app := range apps {
				fmt.Fprintln(p.Output, app.Name)
			}

			return nil
		},
	}

	return apps
}
