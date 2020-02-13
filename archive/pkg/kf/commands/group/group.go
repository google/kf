// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
)

// CommandGroup is a logical grouping of commands
type CommandGroup struct {
	Name     string
	Commands []*cobra.Command
}

// CommandGroups is a list of CommandGroups.
type CommandGroups []CommandGroup

// defaultTemplateFuncs are the func cobra provides its users.
var defaultTemplateFuncs = template.FuncMap{
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"rpad":                    rpad,
	"gt":                      cobra.Gt,
	"eq":                      cobra.Eq,
}

// PrintTemplate prints the provided template to the output writer.
func PrintTemplate(out io.Writer, text string, data interface{}, templateFuncs *template.FuncMap) error {
	t := template.New("")
	if templateFuncs == nil {
		t.Funcs(defaultTemplateFuncs)
	} else {
		t.Funcs(*templateFuncs)
	}

	template.Must(t.Parse(text))
	return t.Execute(out, data)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

/// trimRightSpace trims tailing whitespace.
func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// CalculateMinWidth gets the minimum width required for all command names.
func (groups CommandGroups) CalculateMinWidth() int {
	minWidth := 0
	for _, group := range groups {
		for _, c := range group.Commands {
			if c == nil {
				continue
			}
			if len(c.Name()) > minWidth {
				minWidth = len(c.Name())
			}
		}
	}
	return minWidth
}

// PrintTrimmedMultilineString does just that.
func PrintTrimmedMultilineString(str string, out io.Writer) {
	if str == "" {
		return
	}

	for _, line := range strings.Split(str, "\n") {
		fmt.Fprintln(out, strings.TrimSpace(line))
	}
}

// CommandGroupUsageFunc returns a UsageFunc a root level command can use.
func CommandGroupUsageFunc(groups CommandGroups, templateFuncs *template.FuncMap) func(*cobra.Command) error {
	return func(command *cobra.Command) error {
		out := command.OutOrStdout()
		fmt.Fprintln(out)
		return PrintTemplate(out, command.UsageTemplate(), command, templateFuncs)
	}
}

// CommandGroupHelpFunc returns a HelpFunc a root level command can use.
func CommandGroupHelpFunc(rootCommand *cobra.Command, groups CommandGroups, templateFuncs *template.FuncMap) func(*cobra.Command, []string) {
	return func(command *cobra.Command, args []string) {
		out := command.OutOrStdout()

		// not the root level, use the default template
		if rootCommand != command {
			err := PrintTemplate(out, command.HelpTemplate(), command, templateFuncs)
			if err != nil {
				panic(fmt.Sprintf("Error printing help: %v", err))
			}
			return
		}

		PrintTrimmedMultilineString(command.Long, out)
		fmt.Fprintln(out)

		minWidth := groups.CalculateMinWidth()
		for _, group := range groups {
			fmt.Fprintln(out, group.Name)
			for _, c := range group.Commands {
				fmt.Fprintf(out, "  %s %s\n", rpad(c.Name(), minWidth), c.Short)
			}
			fmt.Fprintln(out)
		}

		fmt.Fprintln(out, "Usage:")
		fmt.Fprintf(out, "  %s [flags] COMMAND\n\n", command.CommandPath())
		fmt.Fprintf(out, "Use \"%s COMMAND --help\" for more information about a given command.\n", command.CommandPath())
	}
}

// AddCommandGroups adds CommandGroups to a given cobra Command.
// A side effact of calling this function is that subcommands will no longer be
// able to use their own usage or help funcs.
func AddCommandGroups(rootCommand *cobra.Command, groups CommandGroups) *cobra.Command {
	if rootCommand == nil {
		panic("nil root command")
	}

	rootCommand.SetUsageFunc(CommandGroupUsageFunc(groups, nil))
	rootCommand.SetHelpFunc(CommandGroupHelpFunc(rootCommand, groups, nil))

	for _, group := range groups {
		rootCommand.AddCommand(group.Commands...)
	}
	return rootCommand
}
