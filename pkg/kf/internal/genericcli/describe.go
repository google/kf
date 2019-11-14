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
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

// NewDescribeCommand creates a describe command.
func NewDescribeCommand(t Type, p *config.KfParams, client dynamic.Interface) *cobra.Command {
	printFlags := genericclioptions.NewPrintFlags("")
	friendlyType := t.FriendlyName()
	commandName := strings.ToLower(friendlyType)

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s NAME", commandName),
		Short:   fmt.Sprintf("Print information about the given %s", friendlyType),
		Long:    fmt.Sprintf("Print information about the given %s", friendlyType),
		Example: fmt.Sprintf("kf %s my-%s", commandName, commandName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if t.Namespaced() {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}
			}

			cmd.SilenceUsage = true

			resourceName := args[0]
			w := cmd.OutOrStdout()

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Getting %s %s in namespace: %s\n", friendlyType, resourceName, p.Namespace); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Getting %s %s\n", friendlyType, resourceName); err != nil {
					return err
				}
			}

			client := getResourceInterface(t, client, p.Namespace)

			resource, err := client.Get(resourceName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if printFlags.OutputFlagSpecified() {
				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}

				// If the type didn't come back with a kind, update it with the
				// type we deserialized it with so the printer will work.
				resource.SetGroupVersionKind(t.GroupVersionKind())
				return printer.PrintObj(resource, w)
			}

			if err := describe.Unstructured(w, resource); err != nil {
				return err
			}

			return nil
		},
	}

	printFlags.AddFlags(cmd)

	// Override output format to be sorted so our generated documents are deterministic
	{
		allowedFormats := printFlags.AllowedFormats()
		sort.Strings(allowedFormats)
		cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(allowedFormats, "|"))
	}

	return cmd
}
