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

package servicebindings

import (
	"fmt"
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/commands/config"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/spf13/cobra"
)

// NewListBindingsCommand allows users to list bindings.
func NewListBindingsCommand(p *config.KfParams, client servicebindings.ClientInterface) *cobra.Command {
	var (
		appName         string
		serviceInstance string
	)

	listCmd := &cobra.Command{
		Use:   "bindings [--app APP_NAME] [--service SERVICE_NAME]",
		Short: "List bindings",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			bindings, err := client.List(
				servicebindings.WithListAppName(appName),
				servicebindings.WithListNamespace(p.Namespace),
				servicebindings.WithListServiceInstance(serviceInstance))
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 1, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "NAME\tAPP\tBINDING NAME\tSERVICE\tSECRET\tREADY\tREASON")
			for _, b := range bindings {
				status := ""
				reason := ""
				for _, cond := range b.Status.Conditions {
					if cond.Type == "Ready" {
						status = fmt.Sprintf("%v", cond.Status)
						reason = cond.Reason
					}
				}
				app := b.Labels[servicebindings.AppNameLabel]
				bindingName := b.Labels[servicebindings.BindingNameLabel]

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s", b.Name, app, bindingName, b.Spec.InstanceRef.Name, b.Spec.SecretName, status, reason)
				fmt.Fprintln(w)
			}

			w.Flush()

			return nil
		},
	}

	listCmd.Flags().StringVarP(
		&appName,
		"app",
		"a",
		"",
		"app to display bindings for")

	listCmd.Flags().StringVarP(
		&serviceInstance,
		"service",
		"s",
		"",
		"service instance to display bindings for")

	return listCmd
}
