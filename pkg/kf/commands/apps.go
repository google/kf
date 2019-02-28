package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewAppsCommand creates a apps command.
func NewAppsCommand(p *KfParams) *cobra.Command {
	var apps = &cobra.Command{
		Use:   "apps",
		Short: "List pushed apps.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			services, err := client.Services(p.Namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}
			services.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"},
			)

			for _, item := range services.Items {
				fmt.Fprintln(p.Output, item.Name)
			}

			return nil
		},
	}

	return apps
}
