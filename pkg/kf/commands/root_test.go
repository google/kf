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
	"context"
	"io"
	"regexp"
	runtime2 "runtime"
	"strings"
	"testing"
	"unicode"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

func TestNewRawKfCommand_style(t *testing.T) {
	t.Parallel()
	root := NewRawKfCommand()

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

			testutil.AssertRegexp(t, "name", "^[a-z][0-9a-z-]+$", cmd.Name())

			testutil.AssertNotBlank(t, "use", cmd.Use)
			testutil.AssertNotBlank(t, "short", cmd.Short)

			if len(cmd.Short) > 80 {
				t.Errorf("Short length is %d, expected <= 80 cols", len(cmd.Short))
			}

			// The root command can behave slightly differently because it needs to
			// describe the tool as opposed to a verb on the tool.
			if cmd.Root() != cmd {
				AssertNotStartWithArticle(t, "short", cmd.Short)

				if !startsWithUpper(cmd.Short) {
					t.Errorf("Short must start with an upper-case character, got: %s", cmd.Short)
				}

				if !endsWithTerminalPuctuation(cmd.Short) {
					t.Errorf("Short must end with terminating punctuation, got: %s", cmd.Short)
				}
			}

			if len(cmd.Commands()) == 0 {
				// Leaf commands need examples.
				testutil.AssertNotBlank(t, "example", cmd.Example)
			} else {
				// Command groups need long descriptions.
				testutil.AssertNotBlank(t, "long", cmd.Long)
			}

			// Any long descriptions that are only one line need to be complete
			// sentences. Multi-line longs do as well, but aren't easy to test.
			if long := cmd.Long; long != "" && !strings.Contains(long, "\n") {
				if !startsWithUpper(long) {
					t.Errorf("Single line long must start with an upper-case character, got: %q", long)
				}

				if !endsWithTerminalPuctuation(long) {
					t.Errorf("Single line long must end with terminating punctuation, got: %q", long)
				}
			}

			// Commands that generate public docs need to have long and example text
			// that can be properly de-indented so it doesn't impact markdown
			// generation. Other commands may use any indentation they need in order
			// to show their text correctly in the CLI.
			if cmd.IsAvailableCommand() && !cmd.IsAdditionalHelpTopicCommand() {
				assertValidHeredoc(t, "Long", cmd.Long)
				assertValidHeredoc(t, "Example", cmd.Example)
			}

			// Any shell comments in examples need to start with a capital.
			// This does not include comments that are appended to the end of
			// a line (e.g., some-example # some-comment).
			hasInvalidComment := regexp.MustCompile(`^\s*\#\s+[a-z]`).MatchString(cmd.Example)
			testutil.AssertFalse(t, "example has invalid shell comment", hasInvalidComment)
		})

		t.Run("flags", func(t *testing.T) {
			if cmd.Hidden {
				t.Skip("Skipping flag test for hidden command")
			}

			cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
				t.Run("flag:"+f.Name, func(t *testing.T) {
					testutil.AssertRegexp(t, "name", "^[a-z][a-z0-9-]+$", f.Name)

					if !startsWithUpper(f.Usage) {
						t.Errorf("usage must start with an upper-case character, got: %s", f.Usage)
					}

					if !endsWithTerminalPuctuation(f.Usage) {
						t.Errorf("usage must end with terminating punctuation, got: %s", f.Usage)
					}

					AssertNotStartWithArticle(t, "usage", f.Usage)
				})
			})
		})

		t.Run("silenceUsage", func(t *testing.T) {
			// All commands should default to silencing their usage. If you
			// want to show to show the usage, you can override it via the
			// RunE function.
			testutil.AssertTrue(t, "silenceUsage", cmd.SilenceUsage)
		})

		for _, sub := range cmd.Commands() {
			checkCommandStyle(t, sub)
		}
	})
}

func assertValidHeredoc(t *testing.T, field string, text string) {
	t.Helper()
	lines := strings.Split(text, "\n")

	// Doc only has one line, no possible issues.
	if len(lines) == 1 {
		return
	}

	leadingLine := lines[0]
	lines = lines[1:]

	// If doc starts with blank line, skip it.
	if leadingLine == "" {
		leadingLine, lines = lines[0], lines[1:]
	}

	// Find first indent.
	indent := ""
	for idx, r := range []rune(leadingLine) {
		if !unicode.IsSpace(r) {
			indent = leadingLine[:idx]
			break
		}
	}

	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, indent) {
			t.Errorf("line %q in field %s has inconsistent heredoc prefix", line, field)
		}
	}
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

func endsWithTerminalPuctuation(text string) bool {
	if len(text) == 0 {
		return false
	}

	switch text[len(text)-1:] {
	case ".", "!", "?":
		return true
	default:
		return false
	}
}

func resetClientFactory(oldFactory func(p *config.KfParams) kubernetes.Interface) {
	k8sClientFactory = oldFactory
}

func get(i kubernetes.Interface) func(*config.KfParams) kubernetes.Interface {
	return func(p *config.KfParams) kubernetes.Interface {
		return i
	}
}

func testcmd(writer io.Writer) *cobra.Command {
	cmd := NewVersionCommand(Version, runtime2.GOOS)
	cmd.SetErr(writer)
	cmd.SetContext(context.Background())
	cmd.Annotations = map[string]string{
		config.SkipVersionCheckAnnotation: "",
	}
	return cmd
}

func TestNotShadowingFlags(t *testing.T) {
	t.Parallel()
	root := NewKfCommand()

	testNotShadowingFlags(t, root)
}

func testNotShadowingFlags(t *testing.T, cmd *cobra.Command) {
	t.Run("command:"+cmd.Name(), func(t *testing.T) {
		// We don't want to propagate upwards.
		flags := make(map[string]bool)

		addFlag := func(flag *pflag.Flag) {
			_, ok := flags["--"+flag.Name]
			testutil.AssertFalse(t, "long redundant "+flag.Name, ok)
			if flag.Shorthand != "" {
				_, ok = flags["-"+flag.Shorthand]
				testutil.AssertFalse(t, "short redundant"+flag.Shorthand, ok)
			}
			flags["--"+flag.Name] = true
			flags["-"+flag.Shorthand] = true
		}

		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			addFlag(flag)
		})

		cmd.VisitParents(func(cmd *cobra.Command) {
			cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
				addFlag(flag)
			})
		})

		for _, cmd := range cmd.Commands() {
			testNotShadowingFlags(t, cmd)
		}
	})
}
