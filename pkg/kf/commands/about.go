package commands

import (
	"fmt"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewAboutCommand creates a command that describes the governing agreement.
func NewAboutCommand() *cobra.Command {
	return &cobra.Command{
		Annotations: map[string]string{
			// don't print the version check when looking up names
			config.SkipVersionCheckAnnotation: "",
		},
		Use:          "about",
		Short:        "Print information about Kf's terms of service.",
		Long:         `Print information about Kf's terms of service.`,
		Example:      "kf about",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), `Copyright 2020 Google LLC

Kf is made available as "Software" under the agreement governing your use of
Google Cloud Platform, including the Service Specific Terms available at
https://cloud.google.com/terms/service-terms.

Kf may only be used on Kubernetes clusters in connection with Google Kubernetes
Engine or Anthos.

Third-party notices for the Kf software can be viewed by running:

    kf third-party-licenses`)
		},
	}
}
