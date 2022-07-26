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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	cliutil "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
)

// NewDescribeCommand creates a describe command.
func NewDescribeCommand(t Type, p *config.KfParams, opts ...DescribeOption) *cobra.Command {
	printFlags := cliutil.NewKfPrintFlags()
	friendlyType := t.FriendlyName()

	options := DescribeOptions{
		WithDescribeCommandName(strings.ToLower(friendlyType)),
	}.Extend(opts)

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s NAME", options.CommandName()),
		Aliases:           options.Aliases(),
		Short:             fmt.Sprintf("Print information about the given %s.", friendlyType),
		Long:              fmt.Sprintf("Print information about the given %s.", friendlyType),
		Example:           fmt.Sprintf("kf %s my-%s", options.CommandName(), options.CommandName()),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: ValidArgsFunction(t, p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if t.Namespaced() {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
			}

			resourceName := args[0]
			w := cmd.OutOrStdout()
			logger := logging.FromContext(ctx)

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				logger.Infof("Getting %s %s in Space: %s", friendlyType, resourceName, p.Space)
			} else {
				logger.Infof("Getting %s %s", friendlyType, resourceName)
			}

			client := GetResourceInterface(ctx, t, dynamicclient.Get(ctx), p.Space)

			resource, err := client.Get(context.Background(), resourceName, metav1.GetOptions{})
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
				resource.SetGroupVersionKind(t.GroupVersionKind(ctx))
				return printer.PrintObj(resource, w)
			}

			describe.Unstructured(w, resource)
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
