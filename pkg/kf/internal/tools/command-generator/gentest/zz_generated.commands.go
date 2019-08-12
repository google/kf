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

package gentest

import (
	"github.com/spf13/cobra"
)

type testCommand struct {
	// flagInt64 stores the value for "--flag-int64"
	flagInt64 int64
	// flagInt64IsSet is set to true if the user sets the flag
	flagInt64IsSet bool
	// flagInt stores the value for "--flag-int"
	flagInt int
	// flagIntIsSet is set to true if the user sets the flag
	flagIntIsSet bool
	// flagFloat64 stores the value for "--flag-float64"
	flagFloat64 float64
	// flagFloat64IsSet is set to true if the user sets the flag
	flagFloat64IsSet bool
	// flagString stores the value for "--flag-string"
	flagString string
	// flagStringIsSet is set to true if the user sets the flag
	flagStringIsSet bool
	// flagBool stores the value for "--flag-bool"
	flagBool bool
	// flagBoolIsSet is set to true if the user sets the flag
	flagBoolIsSet bool
}

// FlagInt64 returns the value for "--flag-int64" and if the user set it.
func (f *testCommand) FlagInt64() (int64, bool) {
	return f.flagInt64, f.flagInt64IsSet
}

// FlagInt returns the value for "--flag-int" and if the user set it.
func (f *testCommand) FlagInt() (int, bool) {
	return f.flagInt, f.flagIntIsSet
}

// FlagFloat64 returns the value for "--flag-float64" and if the user set it.
func (f *testCommand) FlagFloat64() (float64, bool) {
	return f.flagFloat64, f.flagFloat64IsSet
}

// FlagString returns the value for "--flag-string" and if the user set it.
func (f *testCommand) FlagString() (string, bool) {
	return f.flagString, f.flagStringIsSet
}

// FlagBool returns the value for "--flag-bool" and if the user set it.
func (f *testCommand) FlagBool() (bool, bool) {
	return f.flagBool, f.flagBoolIsSet
}

func newTestCommand(run func(x testCommand, cmd *cobra.Command, args []string) error) *cobra.Command {
	x := testCommand{}
	cmd := &cobra.Command{
		Use:     "test-command ARG_NONOPTIONAL OTHER_ARG_NONOPTIONAL [ARG_OPTIONAL]",
		Short:   "some-short",
		Long:    "some-long",
		Example: "  first-example\n  second-example",
		Aliases: []string{"some-alias"},
		Args:    cobra.RangeArgs(2, 3),
	}
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		// FlagInt64 returns the value for "--flag-int64" and if the user set it.

		x.flagInt64IsSet = cmd.Flags().Changed("flag-int64")
		// FlagInt returns the value for "--flag-int" and if the user set it.

		x.flagIntIsSet = cmd.Flags().Changed("flag-int")
		// FlagFloat64 returns the value for "--flag-float64" and if the user set it.

		x.flagFloat64IsSet = cmd.Flags().Changed("flag-float64")
		// FlagString returns the value for "--flag-string" and if the user set it.

		x.flagStringIsSet = cmd.Flags().Changed("flag-string")
		// FlagBool returns the value for "--flag-bool" and if the user set it.

		x.flagBoolIsSet = cmd.Flags().Changed("flag-bool")

	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run(x, cmd, args)
	}

	// FlagInt64 returns the value for "--flag-int64" and if the user set it.
	cmd.Flags().Int64VarP(
		&x.flagInt64,
		"flag-int64",
		"a",
		99,
		"some int64 thing",
	)
	// FlagInt returns the value for "--flag-int" and if the user set it.
	cmd.Flags().IntVarP(
		&x.flagInt,
		"flag-int",
		"b",
		100,
		"some int thing",
	)
	// FlagFloat64 returns the value for "--flag-float64" and if the user set it.
	cmd.Flags().Float64VarP(
		&x.flagFloat64,
		"flag-float64",
		"c",
		101.0,
		"some float64 thing",
	)
	// FlagString returns the value for "--flag-string" and if the user set it.
	cmd.Flags().StringVarP(
		&x.flagString,
		"flag-string",
		"d",
		"one hundred and two",
		"some string thing",
	)
	// FlagBool returns the value for "--flag-bool" and if the user set it.
	cmd.Flags().BoolVarP(
		&x.flagBool,
		"flag-bool",
		"e",
		true,
		"some bool thing",
	)

	return cmd
}

type ManifestTestCommand struct {
	// flagInt64 stores the value for "FlagInt64"
	FlagInt64 int64 `yaml:"flagInt64"`
	// flagInt stores the value for "FlagInt"
	FlagInt int `yaml:"flagInt"`
	// flagFloat64 stores the value for "FlagFloat64"
	FlagFloat64 float64 `yaml:"flagFloat64"`
	// flagString stores the value for "FlagString"
	FlagString string `yaml:"flagString"`
	// flagBool stores the value for "FlagBool"
	FlagBool bool `yaml:"flagBool"`
}
