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

package gentest

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

//go:generate ./generate-test-output.sh

func TestNewCommand_Meta(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand(func(tc testCommand, cmd *cobra.Command, args []string) error {
		return nil
	})
	testutil.AssertEqual(t, "Use", "test-command ARG_NONOPTIONAL OTHER_ARG_NONOPTIONAL [ARG_OPTIONAL]", cmd.Use)
	testutil.AssertEqual(t, "Short", "some-short", cmd.Short)
	testutil.AssertEqual(t, "Long", "some-long", cmd.Long)
	testutil.AssertEqual(t, "Example", []string{"  first-example", "  second-example"}, strings.Split(cmd.Example, "\n"))
	testutil.AssertEqual(t, "Aliases", []string{"some-alias"}, cmd.Aliases)
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand(func(tc testCommand, cmd *cobra.Command, args []string) error {
		return nil
	})
	testutil.AssertErrorsEqual(t, errors.New("accepts between 2 and 3 arg(s), received 0"), cmd.Execute())
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand(func(tc testCommand, cmd *cobra.Command, args []string) error {
		// Int64
		{
			x, ok := tc.FlagInt64()
			testutil.AssertEqual(t, "flag-int64", int64(199), x)
			testutil.AssertEqual(t, "flag-int64 is set", true, ok)
		}
		// Int
		{
			x, ok := tc.FlagInt()
			testutil.AssertEqual(t, "flag-int", 1100, x)
			testutil.AssertEqual(t, "flag-int is set", true, ok)
		}
		// Float64
		{
			x, ok := tc.FlagFloat64()
			testutil.AssertEqual(t, "flag-float64", 1101.0, x)
			testutil.AssertEqual(t, "flag-float64 is set", true, ok)
		}
		// String
		{
			x, ok := tc.FlagString()
			testutil.AssertEqual(t, "flag-string", "one thousand one hundred and two", x)
			testutil.AssertEqual(t, "flag-string is set", true, ok)
		}
		// Bool
		{
			x, ok := tc.FlagBool()
			testutil.AssertEqual(t, "flag-bool", false, x)
			testutil.AssertEqual(t, "flag-bool is set", true, ok)
		}

		return nil
	})
	cmd.SetArgs([]string{
		"--flag-int64=199",
		"--flag-int=1100",
		"--flag-float64=1101.0",
		"--flag-string=one thousand one hundred and two",
		"--flag-bool=false",
		"some-arg-1",
		"some-arg-2",
	})
	testutil.AssertNil(t, "err", cmd.Execute())
}

func TestNewCommand_Defaults(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand(func(tc testCommand, cmd *cobra.Command, args []string) error {
		// Int64
		{
			x, ok := tc.FlagInt64()
			testutil.AssertEqual(t, "flag-int64", int64(99), x)
			testutil.AssertEqual(t, "flag-int64 is set", false, ok)
		}
		// Int
		{
			x, ok := tc.FlagInt()
			testutil.AssertEqual(t, "flag-int", 100, x)
			testutil.AssertEqual(t, "flag-int is set", false, ok)
		}
		// Float64
		{
			x, ok := tc.FlagFloat64()
			testutil.AssertEqual(t, "flag-float64", 101.0, x)
			testutil.AssertEqual(t, "flag-float64 is set", false, ok)
		}
		// String
		{
			x, ok := tc.FlagString()
			testutil.AssertEqual(t, "flag-string", "one hundred and two", x)
			testutil.AssertEqual(t, "flag-string is set", false, ok)
		}
		// Bool
		{
			x, ok := tc.FlagBool()
			testutil.AssertEqual(t, "flag-bool", true, x)
			testutil.AssertEqual(t, "flag-bool is set", false, ok)
		}

		return nil
	})
	cmd.SetArgs([]string{
		"some-arg-1",
		"some-arg-2",
	})
	testutil.AssertNil(t, "err", cmd.Execute())
}

func TestManifest(t *testing.T) {
	t.Parallel()
	var m ManifestTestCommand
	testutil.AssertNil(t, "err", yaml.Unmarshal([]byte(`
flagInt64: 99
flagInt: 100
flagFloat64: 101.0
flagString: one hundred and two
flagBool: true
`), &m))

	testutil.AssertEqual(t, "flagInt64", int64(99), m.FlagInt64)
	testutil.AssertEqual(t, "flagInt", 100, m.FlagInt)
	testutil.AssertEqual(t, "flagFloat64", 101.0, m.FlagFloat64)
	testutil.AssertEqual(t, "flagString", "one hundred and two", m.FlagString)
	testutil.AssertEqual(t, "flagBool", true, m.FlagBool)
}

func ExampleMarkdown() {
	data, err := ioutil.ReadFile("gendocs/markdown.md")
	fmt.Println("error is:", err)
	fmt.Println(string(data))

	// Output: error is: <nil>
	//
	// ---
	// title: "test-command"
	// linkTitle: "test-command"
	// weight: 10
	// ---
	//
	// ### Usage
	// kf test-command ARG_NONOPTIONAL OTHER_ARG_NONOPTIONAL [ARG_OPTIONAL]
	//
	// ### Description
	// some-long
	//
	// ### Aliases
	//
	// * some-alias
	//
	// ### Examples
	//
	//     first-example
	//
	//     second-example
	//
	// ### Positional Arguments
	//
	// ##### arg_nonoptional
	// arg_nonoptional desc (REQUIRED)
	//
	// ##### other_arg_nonoptional
	// other_arg_nonoptional desc (REQUIRED)
	//
	// ##### arg_optional
	// arg_optional desc (OPTIONAL)
	//
	// ### Flags
	//
	// ##### -a, --flag-int64 <_int64_>
	// some int64 thing (default=99)
	//
	// ##### -b, --flag-int <_int_>
	// some int thing (default=100)
	//
	// ##### -c, --flag-float64 <_float64_>
	// some float64 thing (default=101.0)
	//
	// ##### -d, --flag-string <_string_>
	// some string thing (default="one hundred and two")
	//
	// ##### -e, --flag-bool <_bool_>
	// some bool thing (default=true)
}
