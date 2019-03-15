package apps

import (
	"fmt"
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
)

// AppLister lists deployed applications.
type AppLister interface {
	// List the deployed applications in a namespace.
	List(...kf.ListOption) ([]serving.Service, error)
}

// NewAppsCommand creates a apps command.
func NewAppsCommand(p *config.KfParams, l AppLister) *cobra.Command {

	var apps = &cobra.Command{
		Use:   "apps",
		Short: "List pushed apps.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(p.Output, "Getting apps in namespace: %s\n", p.Namespace)
			fmt.Fprintln(p.Output)

			apps, err := l.List(kf.WithListNamespace(p.Namespace))
			if err != nil {
				return err
			}

			// Emulating:
			// https://github.com/knative/serving/blob/master/config/300-service.yaml
			w := tabwriter.NewWriter(p.Output, 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tDOMAIN\tLATESTCREATED\tLATESTREADY\tREADY\tREASON")
			for _, app := range apps {
				status := ""
				reason := ""
				for _, cond := range app.Status.GetConditions() {
					if cond.Type == "Ready" {
						status = fmt.Sprintf("%v", cond.Status)
						reason = cond.Reason
					}
				}

				fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%s\t%s\n",
					app.Name,
					app.Status.Domain,
					app.Status.LatestCreatedRevisionName,
					app.Status.LatestReadyRevisionName,
					status,
					reason)
			}

			w.Flush()

			return nil
		},
	}

	return apps
}
