// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commanddocgenerator

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestPrintFlags(t *testing.T) {
	t.Parallel()

	// variables to be referenced in flags so they parse
	// NOT GUARANTEED to be in any value because tests
	// can run in parallel
	var (
		nopStr    string
		nopStrArr []string
		nopInt    int
		nopBool   bool
	)

	cases := map[string]struct {
		flagSet *pflag.FlagSet
	}{
		"empty": {
			flagSet: (func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				return flags
			}()),
		},
		"flag-formatting": {
			flagSet: (func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

				flags.StringVar(&nopStr, "long-flag", "", "A long flag.")
				flags.StringVarP(&nopStr, "short-long-flag", "s", "", "A long and short flag.")

				return flags
			}()),
		},
		"zero-defaults": {
			flagSet: (func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

				flags.StringVar(&nopStr, "string", "", "A zero string default.")
				flags.IntVar(&nopInt, "int", 0, "A zero int default.")
				flags.StringArrayVar(&nopStrArr, "nilstrarr", nil, "A nil array default.")
				flags.StringArrayVar(&nopStrArr, "strarr", []string{}, "An empty array default.")
				flags.BoolVar(&nopBool, "bool", false, "A zero boolean default.")

				return flags
			}()),
		},
		"nonzero-defaults": {
			flagSet: (func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

				flags.StringVar(&nopStr, "string", "nonzero", "A nonzero string default.")
				flags.IntVar(&nopInt, "int", 10, "A nonzero int default.")
				flags.StringArrayVar(&nopStrArr, "strarr", []string{"foo", "bar"}, "A nonzero array default.")
				flags.BoolVar(&nopBool, "bool", true, "A nonzero boolean default.")

				return flags
			}()),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := &bytes.Buffer{}
			PrintFlags(buf, tc.flagSet)

			testutil.AssertGolden(t, "PrintFlags", buf.Bytes())
		})
	}
}

func TestTraverseCommands(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "root"}
	foo := &cobra.Command{Use: "foo", Run: func(cmd *cobra.Command, args []string) {}}
	bar := &cobra.Command{Use: "bar", Run: func(cmd *cobra.Command, args []string) {}}
	baz := &cobra.Command{Use: "baz", Run: func(cmd *cobra.Command, args []string) {}}
	root.AddCommand(foo, bar)
	bar.AddCommand(baz)

	uses := map[string]bool{}
	f := func(cmd *cobra.Command) error {
		if uses[cmd.Use] {
			t.Fatalf("%s was repeated", cmd.Use)
		}
		uses[cmd.Use] = true
		return nil
	}
	err := TraverseCommands(root, f)
	testutil.AssertNil(t, "err", err)

	expectedUses := map[string]bool{
		"foo": true,
		"bar": true,
		"baz": true,
	}
	testutil.AssertEqual(t, "uses", expectedUses, uses)
}

func TestTraverseCommands_err(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(
		&cobra.Command{Use: "1", Run: func(cmd *cobra.Command, args []string) {}},
		&cobra.Command{Use: "err", Run: func(cmd *cobra.Command, args []string) {}},
		&cobra.Command{Use: "3", Run: func(cmd *cobra.Command, args []string) {}},
	)

	f := func(cmd *cobra.Command) error {
		if cmd.Use == "err" {
			return errors.New("some-error")
		}
		return nil
	}

	err := TraverseCommands(cmd, f)
	testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
}

func TestListHeritage(t *testing.T) {
	t.Parallel()

	barCmd := &cobra.Command{Use: "bar [OPT]"}
	fooCmd := &cobra.Command{Use: "foo"}
	fooCmd.AddCommand(barCmd)

	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(
		fooCmd,
		&cobra.Command{Use: "baz"},
	)

	h := ListHeritage(barCmd)
	testutil.AssertEqual(t, "heritage", []string{"root", "foo", "bar"}, h)
}

func TestGenerateBookYAML(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "root", Run: func(cmd *cobra.Command, args []string) {}}
	foo := &cobra.Command{Use: "foo [BAR]", Run: func(cmd *cobra.Command, args []string) {}}
	bar := &cobra.Command{Use: "bar", Run: func(cmd *cobra.Command, args []string) {}}
	baz := &cobra.Command{Use: "baz", Run: func(cmd *cobra.Command, args []string) {}}
	version := "v2.2.0"
	root.AddCommand(foo, bar)
	bar.AddCommand(baz)

	buf := &bytes.Buffer{}
	GenerateBookYAML(buf, root, version)

	testutil.AssertGolden(t, "GenerateBookYAML", buf.Bytes())
}
