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
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CommandGroup struct {
	Message  string
	Commands []*cobra.Command
}

type CommandGroups []CommandGroup

func SubCommandFlagPrint(command *cobra.Command, out io.Writer) {
	if command.HasLocalFlags() {
		fmt.Fprintf(out, "Flags:\n")
		command.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Name != "help" {
				defaultValue := flag.DefValue
				if defaultValue == "" {
					defaultValue = "''"
				}
				fmt.Fprintf(out, "  --%s=%s: %s\n", flag.Name, defaultValue, flag.Usage)
			}
		})
		fmt.Fprintln(out)
	}
}

func SubCommandUsagePrint(command *cobra.Command, out io.Writer) {

	fmt.Fprintln(out, "Usage:")

	if command.HasParent() {
		fullParentName := GetFullName(command.Parent())
		fmt.Fprintf(out, "  %s %s [flags]\n\n", fullParentName, command.Use)
		fmt.Fprintf(out, "Use \"%s --help\" for a list of global command-line flags.\n", fullParentName)
	} else {
		fmt.Fprintf(out, "  %s [flags]\n\n", command.Use)
	}
}

func CommandGroupUsageFunc(cmd *cobra.Command, groups CommandGroups) func(*cobra.Command) error {
	return func(command *cobra.Command) error {
		out := command.OutOrStdout()
		fmt.Fprintln(out)
		if command.HasExample() {
			fmt.Fprintln(out, "Examples:")
			fmt.Fprintln(out, command.Example)
			fmt.Fprintln(out)
		}
		SubCommandFlagPrint(command, out)
		SubCommandUsagePrint(command, out)
		return nil
	}
}

func CommandGroupHelpFunc(cmd *cobra.Command, groups CommandGroups) func(*cobra.Command, []string) {
	return func(command *cobra.Command, args []string) {
		minWidth := 0
		for _, group := range groups {
			for _, c := range group.Commands {
				if len(c.Name()) > minWidth {
					minWidth = len(c.Name())
				}
			}
		}

		// 2 for the prefix spaces, 1 for the padding
		minWidth += 3

		out := command.OutOrStdout()
		PrintWhitespaceNormalizedString(command.Long, out)
		fmt.Fprintln(out)

		if cmd == command {
			tabout := tabwriter.NewWriter(command.OutOrStdout(), minWidth, 8, 1, ' ', 0)
			defer tabout.Flush()
			for _, group := range groups {
				fmt.Fprintln(tabout, group.Message)
				for _, c := range group.Commands {
					fmt.Fprintf(tabout, "  %s\t%s\n", c.Name(), c.Short)
				}
				fmt.Fprintln(tabout)
			}
			fmt.Fprintln(tabout, "Usage:")
			fmt.Fprintf(tabout, "  %s [flags] COMMAND\n\n", GetFullName(command))
			fmt.Fprintf(tabout, "Use \"%s COMMAND --help\" for more information about a given command.\n", GetFullName(command))
		} else {
			SubCommandFlagPrint(command, out)
			SubCommandUsagePrint(command, out)
		}
	}
}

func PrintWhitespaceNormalizedString(str string, out io.Writer) {
	for _, line := range strings.Split(str, "\n") {
		fmt.Fprintln(out, strings.TrimSpace(line))
	}
}

func GetFullName(command *cobra.Command) string {
	if command == nil {
		return ""
	}

	if command.HasParent() {
		return fmt.Sprintf("%s %s", GetFullName(command.Parent()), command.Name())
	}

	return command.Name()
}

func ActsAsRootCommand(cmd *cobra.Command, groups CommandGroups) *cobra.Command {
	if cmd == nil {
		panic("nil root command")
	}
	cmd.SetUsageFunc(CommandGroupUsageFunc(cmd, groups))
	cmd.SetHelpFunc(CommandGroupHelpFunc(cmd, groups))
	for _, group := range groups {
		for _, command := range group.Commands {
			cmd.AddCommand(command)
		}
	}
	return cmd
}
