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

package genericcli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/internal/tableclient"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

// NewListCommand creates a list command that can print tables.
func NewListCommand(t Type, p *config.KfParams, client dynamic.Interface, tableClient tableclient.Interface) *cobra.Command {
	printFlags := genericclioptions.NewPrintFlags("")
	friendlyType := t.FriendlyName() + "s"
	commandName := strings.ToLower(friendlyType)

	scope := "in the cluster"
	if t.Namespaced() {
		scope = "in the target space"
	}

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s", commandName),
		Short:   fmt.Sprintf("List %s %s", friendlyType, scope),
		Long:    fmt.Sprintf("List %s %s", friendlyType, scope),
		Example: fmt.Sprintf("kf %s", commandName),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if t.Namespaced() {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}
			}

			cmd.SilenceUsage = true
			w := cmd.OutOrStdout()

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				fmt.Fprintf(cmd.ErrOrStderr(), "Listing %s in namespace: %s\n", friendlyType, p.Namespace)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Listing %s\n", friendlyType)
			}

			if printFlags.OutputFlagSpecified() {
				client := getResourceInterface(t, client, p.Namespace)

				resource, err := client.List(metav1.ListOptions{})
				if err != nil {
					return err
				}

				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}

				// If the type didn't come back with a kind, update it with the
				// type we deserialized it with so the printer will work.
				resource.SetGroupVersionKind(t.GroupVersionKind())
				return printer.PrintObj(resource, w)
			}

			table, err := tableClient.Table(t, p.Namespace, metav1.ListOptions{})
			if err != nil {
				return err
			}

			describe.MetaV1Beta1Table(w, table)
			return nil
		},
	}

	printFlags.AddFlags(cmd)

	// Override output format to be sorted so our generated documents are deterministic
	// The following block can be deleted if https://github.com/kubernetes/kubernetes/pull/82836
	// gets merged.
	{
		allowedFormats := printFlags.AllowedFormats()
		sort.Strings(allowedFormats)
		cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(allowedFormats, "|"))
	}

	return cmd
}
