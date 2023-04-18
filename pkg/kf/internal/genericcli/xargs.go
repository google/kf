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
	"html/template"
	"os/exec"
	"strings"

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

	scope := "in the cluster"
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

	var example string
	if options.Example() != "" {
		example = options.Example()
	} else {
		example = fmt.Sprintf("kf %s [flags]", options.CommandName())
	}

	var short string
	if options.Short() != "" {
		short = options.Short()
	} else {
		short = fmt.Sprintf("xargs %s %s", options.PluralFriendlyName(), scope)
	}

	var long string
	if options.Long() != "" {
		long = options.Long()
	} else {
		long = fmt.Sprintf("xargs runs a command on each %s %s.", t.FriendlyName(), scope)
	}

	cmd := &cobra.Command{
		Use:          strings.TrimSpace(fmt.Sprintf("%s %s", options.CommandName(), argNames)),
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

			if flags.allSpaces {
				logger.Infof("Xargs %s in all spaces", options.PluralFriendlyName())
			} else if t.Namespaced() {
				logger.Infof("Xargs %s in space: %s", options.PluralFriendlyName(), p.Space)
			}

			commandTmpl := strings.Join(args, " ")
			tmpl, err := template.New("command").Parse(commandTmpl)
			if err != nil {
				return fmt.Errorf("couldn't parse command template %s: %v", commandTmpl, err)
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

			group, groupCtx := errgroup.WithContext(ctx)
			limit := semaphore.NewWeighted(int64(flags.spaceConcurrency))

			for _, space := range spaces {
				limit.Acquire(groupCtx, 1)
				if groupCtx.Err() != nil {
					break
				}

				// Shadow space as a local to be captured by the closure
				space := space
				group.Go(func() error {
					defer limit.Release(1)
					return doSpace(groupCtx, cmd, t, listOptions, flags, space, tmpl)
				})
			}

			if err := group.Wait(); err != nil {
				return err
			}

			if flags.dryRun {
				fmt.Fprintf(
					cmd.OutOrStdout(), color.New(color.FgHiYellow).Sprint("Run with --dry-run=false to apply."))
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	for flag, label := range options.LabelFilters() {
		usage := fmt.Sprintf("Sets a filter for the %q label.", label)
		labelSelectorFlags[label] = cmd.Flags().String(flag, "", usage)
	}

	flags.Add(cmd)

	return cmd
}

type XargFlags struct {
	dryRun           bool
	allSpaces        bool
	spaceConcurrency int
	appConcurrency   int
}

func (f *XargFlags) Add(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", true, "Enables dry-run mode, commands are printed but will not be executed.")
	cmd.Flags().BoolVar(&f.allSpaces, "all-spaces", false, "Enables targeting all spaces in the cluster.")
	cmd.Flags().IntVar(&f.spaceConcurrency, "space-concurrency", 1, "The number of spaces that may be operated on in parallel. Only meaningful when multiple spaces are targeted.")
	cmd.Flags().IntVar(&f.appConcurrency, "app-concurrency", 1, "The number of apps within a space that may be operated on in parallel. Note that total possible concurrency is space-concurrency * app-concurrency.")
}

func doSpace(
	ctx context.Context,
	cmd *cobra.Command,
	t Type,
	listOptions metav1.ListOptions,
	flags *XargFlags,
	space string,
	commandTemplate *template.Template,
) error {
	client := GetResourceInterface(ctx, t, dynamicclient.Get(ctx), space)
	resource, err := client.List(ctx, listOptions)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	limit := semaphore.NewWeighted(int64(flags.appConcurrency))

	for _, item := range resource.Items {
		limit.Acquire(ctx, 1)
		if ctx.Err() != nil {
			break
		}

		// Shadow item as a local variable to be captured by the closure
		item := item
		group.Go(func() error {
			b := bytes.NewBuffer([]byte{})
			err = commandTemplate.Execute(b, &struct {
				Name  string
				Space string
			}{
				Name:  item.GetName(),
				Space: item.GetNamespace(),
			})
			if err != nil {
				return err
			}

			c := b.String()

			w := textio.NewPrefixWriter(cmd.OutOrStdout(), color.New(color.FgGreen).Sprintf("[space=%s app=%s] ", space, item.GetName()))
			fmt.Fprintf(w, "Exec: %v\n", c)

			if flags.dryRun {
				return nil
			}

			p := exec.Command("bash", "-c", c)
			p.Stdout = w
			p.Stderr = w
			if err := p.Start(); err != nil {
				return err
			}

			if err := p.Wait(); err != nil {
				fmt.Fprintf(w, color.New(color.FgHiRed).Sprintf("Command '%v' failed: %v", c, err))
				fmt.Fprintf(w, color.New(color.Reset).Sprintln())
				return err
			}
			return nil
		})
	}

	return group.Wait()
}
