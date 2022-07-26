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

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/kf/injection/clients/tableclient"
	cliutil "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
)

// ListArgumentFilterHandler is a callback for a CLI argument that modifies
// a Kubernetes ListOptions selector.
type ListArgumentFilterHandler func(argValue string, opts *metav1.ListOptions) error

// ListArgumentFilter allows attaching optional filters to arguments.
type ListArgumentFilter struct {
	// Name contains the name of the argument.
	Name string
	// Handler is a callback for the argument.
	Handler ListArgumentFilterHandler
	// Required indicates if the argument is required.
	Required bool
}

// NewAddLabelFilter creates a ListArgumentFilterHandler that adds an additional
// label to the existing selector.
func NewAddLabelFilter(labelKey string) ListArgumentFilterHandler {
	return func(argValue string, opts *metav1.ListOptions) error {
		parsed, err := labels.Parse(opts.LabelSelector)
		if err != nil {
			return err
		}

		requirement, err := labels.NewRequirement(labelKey, selection.Equals, []string{argValue})
		if err != nil {
			return err
		}

		opts.LabelSelector = parsed.Add(*requirement).String()
		return nil
	}
}

// NewListCommand creates a list command that can print tables.
func NewListCommand(t Type, p *config.KfParams, opts ...ListOption) *cobra.Command {
	printFlags := cliutil.NewKfPrintFlags()

	options := ListOptions{
		WithListPluralFriendlyName(t.FriendlyName() + "s"),
		WithListCommandName(strings.ToLower(t.FriendlyName() + "s")),
	}.Extend(opts)

	scope := "in the cluster"
	if t.Namespaced() {
		scope = "in the targeted Space"
	}

	labelSelectorFlags := map[string]*string{}

	var argNames string
	var requiredArgs int
	var maxArgs int
	argFilters := options.ArgumentFilters()
	for i := len(argFilters) - 1; i >= 0; i-- {
		argNames = strings.TrimSpace(fmt.Sprintf("%s %s", argFilters[i].Name, argNames))
		maxArgs++
		if argFilters[i].Required {
			argNames = fmt.Sprintf("%s", argNames)
			requiredArgs++
		} else {
			argNames = fmt.Sprintf("[%s]", argNames)
		}
	}

	var args cobra.PositionalArgs
	if maxArgs == requiredArgs {
		args = cobra.ExactArgs(requiredArgs)
	} else {
		args = cobra.RangeArgs(requiredArgs, maxArgs)
	}

	var example string
	if options.Example() != "" {
		example = options.Example()
	} else {
		example = fmt.Sprintf("kf %s", options.CommandName())
	}

	var short string
	if options.Short() != "" {
		short = options.Short()
	} else {
		short = fmt.Sprintf("List %s %s.", options.PluralFriendlyName(), scope)
	}

	var long string
	if options.Long() != "" {
		long = options.Long()
	} else {
		long = fmt.Sprintf("List %s %s.", options.PluralFriendlyName(), scope)
	}

	cmd := &cobra.Command{
		Use:          strings.TrimSpace(fmt.Sprintf("%s %s", options.CommandName(), argNames)),
		Aliases:      options.Aliases(),
		Short:        short,
		Long:         long,
		Example:      example,
		Args:         args,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if t.Namespaced() {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
			}

			w := cmd.OutOrStdout()
			logger := logging.FromContext(ctx)

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				logger.Infof("Listing %s in Space: %s", options.PluralFriendlyName(), p.Space)
			} else {
				logger.Infof("Listing %s", options.PluralFriendlyName())
			}

			labelSelector := labels.Set{}
			for k, v := range labelSelectorFlags {
				if v != nil && *v != "" {
					labelSelector[k] = *v
				}
			}

			listOptions := metav1.ListOptions{
				LabelSelector: labels.
					SelectorFromSet(labelSelector).
					Add(options.LabelRequirements()...).
					String(),
			}

			for i, arg := range args {
				if err := argFilters[i].Handler(arg, &listOptions); err != nil {
					return fmt.Errorf("couldn't parse argument %d value: %q: %v", i+1, arg, err)
				}
			}

			if printFlags.OutputFlagSpecified() {
				client := GetResourceInterface(ctx, t, dynamicclient.Get(ctx), p.Space)

				resource, err := client.List(ctx, listOptions)
				if err != nil {
					return err
				}

				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}

				// If the type didn't come back with a kind, update it with the
				// type we deserialized it with so the printer will work.
				resource.SetGroupVersionKind(t.GroupVersionKind(ctx))
				return printer.PrintObj(resource, w)
			}

			table, err := tableclient.Get(ctx).Table(ctx, t, p.Space, listOptions)
			if err != nil {
				return err
			}
			describe.MetaV1Beta1Table(w, table)
			return nil
		},
	}

	for flag, label := range options.LabelFilters() {
		usage := fmt.Sprintf("Sets a filter for the %q label.", label)
		labelSelectorFlags[label] = cmd.Flags().String(flag, "", usage)
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
