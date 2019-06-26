// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apps

import (
	"fmt"
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewAppsCommand creates a apps command.
func NewAppsCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	var apps = &cobra.Command{
		Use:     "apps",
		Short:   "List pushed apps",
		Example: `  kf apps`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Getting apps in namespace: %s\n", p.Namespace)

			apps, err := appsClient.List(p.Namespace)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Found %d apps in namespace %s\n", len(apps), p.Namespace)
			fmt.Fprintln(cmd.OutOrStdout())

			// Emulating:
			// https://github.com/knative/serving/blob/master/config/300-service.yaml
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tDOMAIN\tLATEST CREATED\tLATEST READY\tREADY\tREASON")
			for _, app := range apps {
				var status, reason string
				if cond := app.Status.GetCondition("Ready"); cond != nil {
					status = fmt.Sprintf("%v", cond.Status)
					reason = cond.Reason
				}

				if !app.DeletionTimestamp.IsZero() {
					reason = "Deleting"
				}

				if app.Name == "" {
					continue
				}

				host := ""
				if app.Status.Address != nil {
					host = app.Status.Address.Hostname
				}

				fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%s\t%s\n",
					app.Name,
					host,
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
