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

package completion

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// bashCompletionFuncName gets the name of a bash completion func for the given
// type.
func bashCompletionFuncName(k8sType string) string {
	return fmt.Sprintf("__kf_name_%s", k8sType)
}

// bashCompletionFunc returns the bash completion function for a single type.
func bashCompletionFunc(k8sType string) string {
	return bashCompletionFuncName(k8sType) + `()
{
  local out
  if out=$(kf names ` + k8sType + ` 2>/dev/null); then
      COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
  fi
}
`
}

// MarkFlagCompletionSupported adds a completion annotation to a flag.
func MarkFlagCompletionSupported(flags *pflag.FlagSet, name, k8sType string) error {
	return flags.SetAnnotation(name, cobra.BashCompCustom, []string{bashCompletionFuncName(k8sType)})
}

// MarkArgCompletionSupported returns completion annotations for a CobraCommand
func MarkArgCompletionSupported(cmd *cobra.Command, k8sType string) {
	if cmd == nil {
		return
	}

	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations[cobra.BashCompCustom] = bashCompletionFuncName(k8sType)
}

func customCompletions(cmd *cobra.Command) map[string]string {
	out := make(map[string]string)

	if customFunc, ok := cmd.Annotations[cobra.BashCompCustom]; ok {
		// Copied from Cobra's path to bash generator
		commandName := cmd.CommandPath()
		commandName = strings.Replace(commandName, " ", "_", -1)
		commandName = strings.Replace(commandName, ":", "__", -1)

		out[commandName] = customFunc
	}

	for _, c := range cmd.Commands() {
		childrenCompletions := customCompletions(c)
		for k, v := range childrenCompletions {
			out[k] = v
		}
	}

	return out
}

// AddBashCompletion adds bash completion to the given Cobra command.
func AddBashCompletion(rootCommand *cobra.Command) error {
	out := &bytes.Buffer{}

	for _, k8sType := range KnownGenericTypes() {
		if _, err := fmt.Fprintln(out, bashCompletionFunc(k8sType)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "__kf_custom_func() {"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "    case ${last_command} in"); err != nil {
		return err
	}

	for commandName, completionFunc := range customCompletions(rootCommand) {
		if _, err := fmt.Fprintf(out, "    %s)\n", commandName); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "        %s\n", completionFunc); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, "        return"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, "        ;;"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(out, "    *)"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "        ;;"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "    esac"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "}"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	rootCommand.BashCompletionFunction = out.String()

	return nil
}
