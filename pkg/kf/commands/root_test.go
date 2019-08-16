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

package commands

import (
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func TestNewKfCommand_style(t *testing.T) {
	root := NewKfCommand()

	checkCommandStyle(t, root)
}

func checkCommandStyle(t *testing.T, cmd *cobra.Command) {
	if cmd.Hidden {
		return
	}

	t.Run("command:"+cmd.Name(), func(t *testing.T) {

		t.Run("docs", func(t *testing.T) {
			if cmd.Hidden {
				t.Skip("Skipping docs test for hidden command")
			}

			testutil.AssertRegexp(t, "name", "^[a-z][a-z-]+$", cmd.Name())

			testutil.AssertNotBlank(t, "use", cmd.Use)
			testutil.AssertNotBlank(t, "short", cmd.Short)

			if len(cmd.Short) > 80 {
				t.Errorf("Short length is %d, expected <= 80 cols", len(cmd.Short))
			}

			if !startsWithUpper(cmd.Short) {
				t.Errorf("Short must start with an upper-case character, got: %s", cmd.Short)
			}

			if len(cmd.Commands()) == 0 {
				// leaf commands need examples
				testutil.AssertNotBlank(t, "example", cmd.Example)
			} else {
				// command groups need long descriptions
				testutil.AssertNotBlank(t, "long", cmd.Long)
			}
		})

		t.Run("flags", func(t *testing.T) {
			if cmd.Hidden {
				t.Skip("Skipping flag test for hidden command")
			}

			cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
				t.Run("flag:"+f.Name, func(t *testing.T) {
					testutil.AssertRegexp(t, "name", "^[a-z][a-z-]+$", f.Name)

					if !startsWithUpper(f.Usage) {
						t.Errorf("usage must start with an upper-case character, got: %s", f.Usage)
					}

					AssertNotStartWithArticle(t, "usage", f.Usage)
				})
			})
		})

		// TODO test parsing examples
		// for each example
		// if leaf, ensure Find() finds this command
		// if node, ensure Find() passes this node
		// ParseFlags

		for _, sub := range cmd.Commands() {
			checkCommandStyle(t, sub)
		}
	})
}

func AssertNotStartWith(t *testing.T, field string, text string, illegal ...string) {
	t.Helper()

	firstWord := strings.ToUpper(strings.Split(text, " ")[0])

	for _, word := range illegal {
		if strings.ToUpper(word) == firstWord {
			t.Errorf("%s must not start with %q", field, word)

		}
	}
}

func AssertNotStartWithArticle(t *testing.T, field string, text string) {
	t.Helper()

	AssertNotStartWith(t, field, text, "the", "a")
}

func startsWithUpper(text string) bool {
	if len(text) == 0 {
		return false
	}

	return strings.ToUpper(text[0:1]) == text[0:1]
}
