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

package cli

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Runner is invoked if the InteractiveNode comes up in the decision digraph.
// What it returns guides the interactive experience.
type Runner func(context.Context, *cobra.Command, []string) (*InteractiveNode, error)

// InteractiveNode is a node of the decision digraph.
type InteractiveNode struct {
	// Setup is ran while traversing the digraph. Its Runner is only invoked
	// if it comes up in the decision digraph. The returned InteractiveNodes
	// are ALL the possible paths the interactive tutorial can go.
	Setup func(*pflag.FlagSet) ([]*InteractiveNode, Runner)

	// ctx gets set by InteractiveWithContext(). If it's not set, the normal
	// context is used during execution.
	ctx context.Context
}

// InteractiveWithContext sets a context for the node to use.
func InteractiveWithContext(ctx context.Context, n *InteractiveNode) *InteractiveNode {
	n.ctx = ctx
	return n
}

// CommandInfo stores the information about a command.
type CommandInfo struct {
	// Use is the one-line usage message.
	Use string

	// Short is the short description shown in the 'help' output.
	Short string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	// Example is examples of how to use the command.
	Example string

	// LogPrefix is a non-cobra command property that sets the prefix for the
	// logger.
	LogPrefix string
}

type showUsageErr struct {
	error
}

// ShowUsage returns an error that implies that the Command should show the
// usage when it occurs.  If this error type is not returned, the usage will
// not be displayed.
func ShowUsage(err error) error {
	return showUsageErr{error: err}
}

// ClearDefaultsForInteractive will clear any defaulted values if the user did
// not set them but asked for interactive mode. This is useful to allow
// commands to set defaults for non-interactive mode, but still allow the user
// to select something in interactive mode.
//
// You can also pass in a blacklist of flags you don't want cleared.
func ClearDefaultsForInteractive(ctx context.Context, flags *pflag.FlagSet, blacklist ...string) {
	if !GetInteractiveMode(ctx) {
		// Non-interactive mode, don't clear anything
		return
	}

	blacklistM := map[string]bool{}
	for _, b := range blacklist {
		blacklistM[b] = true
	}
	flags.VisitAll(func(f *pflag.Flag) {
		if blacklistM[f.Name] || f.Changed {
			return
		}

		// Clear the value
		f.Value.Set("")
	})
}

// NewInteractiveCommand traverses the digraph starting with the given node.
// Each node's flags are added.
func NewInteractiveCommand(
	info CommandInfo,
	root *InteractiveNode,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     info.Use,
		Short:   info.Short,
		Long:    info.Long,
		Example: info.Example,
	}

	verbose := cmd.Flags().BoolP("verbose", "v", false, "Make the operation more chatty")
	interactiveMode := cmd.Flags().Bool("interactive", false, "Make the command interactive")

	m := map[*InteractiveNode]Runner{}

	traverseNodes(root, m, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := SetContextOutput(context.Background(), cmd.ErrOrStderr())
		ctx = SetLogPrefix(ctx, info.LogPrefix)
		ctx = SetVerbosity(ctx, *verbose)
		ctx = SetInteractiveMode(ctx, *interactiveMode)

		err := executeDigraph(ctx, root, m, cmd, args)
		_, showUsage := err.(showUsageErr)
		cmd.SilenceUsage = !showUsage
		return err
	}

	return cmd
}

func traverseNodes(
	n *InteractiveNode,
	history map[*InteractiveNode]Runner,
	cmd *cobra.Command,
) {
	if history[n] != nil {
		return
	}

	childNodes, runner := n.Setup(cmd.Flags())
	history[n] = runner

	for _, n := range childNodes {
		traverseNodes(n, history, cmd)
	}

	return
}

func executeDigraph(
	ctx context.Context,
	n *InteractiveNode,
	nodes map[*InteractiveNode]Runner,
	cmd *cobra.Command,
	args []string,
) error {
	// A nil node implies we've reached the end of the decision digraph.
	if n == nil {
		return nil
	}

	r := nodes[n]
	if r == nil {
		return errors.New("interactive error: unexpected next node")
	}

	if n.ctx != nil {
		ctx = n.ctx
	}

	nextNode, err := r(ctx, cmd, args)
	if err != nil {
		return err
	}

	return executeDigraph(ctx, nextNode, nodes, cmd, args)
}
