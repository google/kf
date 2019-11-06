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

package cli_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/cli"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestNewInteractiveCommand_info(t *testing.T) {
	t.Parallel()
	cmd := cli.NewInteractiveCommand(cli.CommandInfo{
		Use:     "some-usage",
		Short:   "some-short",
		Long:    "some-long",
		Example: "some-example",
	}, &cli.InteractiveNode{
		Setup: func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
			return nil,
				func(ctx context.Context, cmd *cobra.Command, args []string) (*cli.InteractiveNode, error) {
					return nil, nil
				}
		},
	})
	testutil.AssertEqual(t, "Use", "some-usage", cmd.Use)
	testutil.AssertEqual(t, "Short", "some-short", cmd.Short)
	testutil.AssertEqual(t, "Long", "some-long", cmd.Long)
	testutil.AssertEqual(t, "Example", "some-example", cmd.Example)
}

func TestNewInteractiveCommand_traverses_each(t *testing.T) {
	t.Parallel()

	g, _ := buildGraph(t, nil)
	cmd := cli.NewInteractiveCommand(cli.CommandInfo{}, g)

	for _, flagName := range []string{"A", "B", "C"} {
		testutil.AssertNotNil(t, flagName, cmd.Flags().Lookup(flagName))
	}
}

func TestNewInteractiveCommand_execute_each(t *testing.T) {
	t.Parallel()

	g, after := buildGraph(t, nil)
	cmd := cli.NewInteractiveCommand(cli.CommandInfo{}, g)
	testutil.AssertNil(t, "err", cmd.Execute())
	after(t)
}

func TestNewInteractiveCommand_execute_error(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		err          error
		silenceUsage bool
	}{
		"show usage": {
			err:          cli.ShowUsage(errors.New("some-error")),
			silenceUsage: false,
		},
		"silence usage": {
			err:          errors.New("some-error"),
			silenceUsage: true,
		},
	} {
		t.Run(tn, func(*testing.T) {
			g, after := buildGraph(t, tc.err)
			cmd := cli.NewInteractiveCommand(cli.CommandInfo{}, g)
			testutil.AssertNotNil(t, "err", cmd.Execute())
			testutil.AssertEqual(t, "SilenceUsage", tc.silenceUsage, cmd.SilenceUsage)
			after(t)
		})
	}
}

func TestNewInteractiveCommand_invalid_graph(t *testing.T) {
	t.Parallel()

	cmd := cli.NewInteractiveCommand(
		cli.CommandInfo{},
		&cli.InteractiveNode{
			Setup: func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
				return nil,
					func(ctx context.Context, cmd *cobra.Command, args []string) (*cli.InteractiveNode, error) {
						// Invalid because we didn't say had this node as a
						// child.
						return &cli.InteractiveNode{}, nil
					}
			},
		})

	testutil.AssertErrorsEqual(
		t,
		errors.New("interactive error: unexpected next node"),
		cmd.Execute(),
	)
}

func buildGraph(t *testing.T, err error) (*cli.InteractiveNode, func(t *testing.T)) {
	var a, b, c string
	var ia, ib, ic cli.InteractiveNode

	var idx int

	setupC := func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
		flags.StringVar(&c, "C", "", "c flag")
		// Link back to a to create a cycle
		return []*cli.InteractiveNode{&ib, &ia},
			func(ctx context.Context, cmd *cobra.Command, args []string) (*cli.InteractiveNode, error) {
				idx++
				testutil.AssertEqual(t, "C", 3, idx)
				return nil, nil
			}
	}

	setupB := func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
		flags.StringVar(&b, "B", "", "b flag")
		return []*cli.InteractiveNode{&ic},
			func(ctx context.Context, cmd *cobra.Command, args []string) (*cli.InteractiveNode, error) {
				idx++
				testutil.AssertEqual(t, "B", 2, idx)
				return &ic, err
			}
	}

	setupA := func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
		flags.StringVar(&a, "A", "", "a flag")
		return []*cli.InteractiveNode{&ib, &ic},
			func(ctx context.Context, cmd *cobra.Command, args []string) (*cli.InteractiveNode, error) {
				idx++
				testutil.AssertEqual(t, "A", 1, idx)
				return &ib, nil
			}
	}

	ia.Setup = setupA
	ib.Setup = setupB
	ic.Setup = setupC

	return &ia, func(t *testing.T) {
		if err != nil {
			testutil.AssertEqual(t, "Calls", 2, idx)
			return
		}
		testutil.AssertEqual(t, "Calls", 3, idx)
	}
}
