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
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/segmentio/textio"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
)

// NewListCommand creates a list command that can print tables.
func NewXargsCommand(t Type, p *config.KfParams, spacesClient spaces.Client, opts ...XargsOption) *cobra.Command {
	options := XargsOptions{
		WithXargsPluralFriendlyName(t.FriendlyName() + "s"),
		WithXargsCommandName(fmt.Sprintf("xargs-%s", strings.ToLower(t.FriendlyName()+"s"))),
	}.Extend(opts)

	flags := &XargFlags{}

	labelSelectorFlags := map[string]*string{}

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
		short = fmt.Sprintf("Run a command for every %s.", t.FriendlyName())
	}

	var long string
	if options.Long() != "" {
		long = options.Long()
	} else if t.Namespaced() {
		long = fmt.Sprintf("Run a command for every %s in targeted spaces.", t.FriendlyName())
	} else {
		long = fmt.Sprintf("Run a command for every %s.", t.FriendlyName())
	}

	cmd := &cobra.Command{
		Use:          strings.TrimSpace(fmt.Sprintf("%s", options.CommandName())),
		Aliases:      options.Aliases(),
		Short:        short,
		Long:         long,
		Example:      example,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			logger := logging.FromContext(ctx)

			// In dryrun mode disable concurrency for consistent print ordering
			if flags.dryRun {
				flags.resourceConcurrency = 1
				flags.spaceConcurrency = 1
			}

			if flags.allSpaces {
				logger.Infof("# Xargs %s in all spaces", options.PluralFriendlyName())
			} else if t.Namespaced() {
				logger.Infof("# Xargs %s in space: %s", options.PluralFriendlyName(), p.Space)
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

			// Collect spaces to target
			var spaces []string
			if t.Namespaced() {
				if flags.allSpaces {
					spacesList, err := spacesClient.List(ctx)
					if err != nil {
						return err
					}
					for _, space := range spacesList {
						spaces = append(spaces, space.Name)
					}
				} else {
					spaces = strings.Split(p.Space, ",")
				}
			} else {
				spaces = []string{""}
			}

			concurrency := flags.spaceConcurrency
			if concurrency == -1 {
				concurrency = math.MaxInt
			}
			group := new(errgroup.Group)
			limit := semaphore.NewWeighted(int64(concurrency))

			for _, space := range spaces {
				limit.Acquire(ctx, 1)

				// Shadow space as a local to be captured by the closure
				space := space
				group.Go(func() error {
					defer limit.Release(1)
					return doSpace(ctx, cmd, t, listOptions, flags, space, args)
				})
			}

			if err := group.Wait(); err != nil {
				return err
			}

			if flags.dryRun && !flags.generateScript {
				fmt.Fprintf(
					cmd.OutOrStdout(), color.New(color.FgHiYellow).Sprint("# Run with --dry-run=false to apply."))
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	for flag, label := range options.LabelFilters() {
		usage := fmt.Sprintf("Sets a filter for the %q label.", label)
		labelSelectorFlags[label] = cmd.Flags().String(flag, "", usage)
	}

	flags.Add(cmd, t)

	return cmd
}

type XargFlags struct {
	dryRun              bool
	allSpaces           bool
	spaceConcurrency    int
	resourceConcurrency int
	generateScript      bool
}

func (f *XargFlags) Add(cmd *cobra.Command, t Type) {
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", true, "Enables dry-run mode, commands are printed but will not be executed.")
	cmd.Flags().IntVar(&f.resourceConcurrency, "resource-concurrency", 1, "Number of apps within a space that may be operated on in parallel. Total concurrency will be upto space-concurrency * app-concurrency. -1 for no limit.")
	if t.Namespaced() {
		cmd.Flags().BoolVar(&f.allSpaces, "all-namespaces", false, "Enables targeting all spaces in the cluster.")
		cmd.Flags().IntVar(&f.spaceConcurrency, "space-concurrency", -1, "Number of spaces that may be operated on in parallel. -1 for no limit.")
	}
}

func doSpace(
	ctx context.Context,
	cmd *cobra.Command,
	t Type,
	listOptions metav1.ListOptions,
	flags *XargFlags,
	space string,
	args []string,
) error {
	client := GetResourceInterface(ctx, t, dynamicclient.Get(ctx), space)
	resource, err := client.List(ctx, listOptions)
	if err != nil {
		return err
	}

	concurrency := flags.resourceConcurrency
	if concurrency == -1 {
		concurrency = math.MaxInt
	}

	group := new(errgroup.Group)
	limit := semaphore.NewWeighted(int64(concurrency))

	for _, item := range resource.Items {
		limit.Acquire(ctx, 1)

		// Shadow item as a local variable to be captured by the closure
		item := item
		group.Go(func() error {
			defer limit.Release(1)
			c, err := generateCommand(space, item.GetName(), args)
			if err != nil {
				return err
			}

			var w io.Writer = cmd.OutOrStdout()
			if !flags.dryRun {
				w = textio.NewPrefixWriter(w, color.New(color.FgGreen).Sprintf("[space=%s %s=%s] ", space, t.FriendlyName(), item.GetName()))
			}

			if t.Namespaced() {
				fmt.Fprintf(w, "# Command for space=%s %s=%s\n", item.GetNamespace(), t.FriendlyName(), item.GetName())
			} else {
				fmt.Fprintf(w, "# Command for %s=%s\n", t.FriendlyName(), item.GetName())
			}

			if flags.dryRun {
				fmt.Fprintf(w, "%v\n", shellescape.QuoteCommand(c))
				return nil
			}

			p := exec.Command(c[0], c[1:]...)
			p.Stdout = w
			p.Stderr = w
			if err := p.Start(); err != nil {
				return err
			}

			if err := p.Wait(); err != nil {
				fmt.Fprintf(w, color.New(color.FgHiRed).Sprintf("Command '%v' failed: %v", shellescape.QuoteCommand(c), err))
				fmt.Fprintf(w, color.New(color.Reset).Sprintln())
				return err
			}
			return nil
		})
	}

	return group.Wait()
}

func generateCommand(space string, resource string, args []string) ([]string, error) {
	cmd := []string{}
	for _, arg := range args {
		t, err := template.New("").Parse(arg)
		if err != nil {
			return nil, err
		}
		b := bytes.NewBuffer([]byte{})
		err = t.Execute(b, &struct {
			Name  string
			Space string
		}{
			Name:  resource,
			Space: space,
		})
		if err != nil {
			return nil, err
		}
		cmd = append(cmd, b.String())
	}
	return cmd, nil
}
