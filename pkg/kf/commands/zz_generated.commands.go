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

// This file was generated with command-generator.go, DO NOT EDIT IT.

package commands

import (
	"github.com/spf13/cobra"
)

type debug struct {
}

func newDebug(run func(x debug, cmd *cobra.Command, args []string) error) *cobra.Command {
	x := debug{}
	cmd := &cobra.Command{
		Use:     "debug ",
		Short:   "Show debugging information useful for filing a bug report",
		Long:    "",
		Example: "  kf debug",
		Aliases: []string{},
		Args:    cobra.RangeArgs(0, 0),
	}
	cmd.PreRun = func(cmd *cobra.Command, args []string) {

	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run(x, cmd, args)
	}

	return cmd
}

type target struct {
	// space stores the value for "--space"
	space string
	// spaceIsSet is set to true if the user sets the flag
	spaceIsSet bool
}

// Space returns the value for "--space" and if the user set it.
func (f *target) Space() (string, bool) {
	return f.space, f.spaceIsSet
}

func newTarget(run func(x target, cmd *cobra.Command, args []string) error) *cobra.Command {
	x := target{}
	cmd := &cobra.Command{
		Use:     "target ",
		Short:   "Set or view the targeted space",
		Long:    "",
		Example: "  kf target\n  kf target -s myspace",
		Aliases: []string{},
		Args:    cobra.RangeArgs(0, 0),
	}
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Space returns the value for "--space" and if the user set it.

		x.spaceIsSet = cmd.Flags().Changed("space")

	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run(x, cmd, args)
	}

	// Space returns the value for "--space" and if the user set it.
	cmd.Flags().StringVarP(
		&x.space,
		"space",
		"s",
		"",
		"Target the given space.",
	)

	return cmd
}

type version struct {
}

func newVersion(run func(x version, cmd *cobra.Command, args []string) error) *cobra.Command {
	x := version{}
	cmd := &cobra.Command{
		Use:     "version ",
		Short:   "Display the CLI version",
		Long:    "",
		Example: "  kf version",
		Aliases: []string{},
		Args:    cobra.RangeArgs(0, 0),
	}
	cmd.PreRun = func(cmd *cobra.Command, args []string) {

	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run(x, cmd, args)
	}

	return cmd
}
