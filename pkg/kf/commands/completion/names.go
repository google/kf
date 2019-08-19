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

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var namespacedTypes = map[string]schema.GroupVersionResource{
	"apps": schema.GroupVersionResource{
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "apps",
	},
}

var globalTypes = map[string]schema.GroupVersionResource{
	"spaces": schema.GroupVersionResource{
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "spaces",
	},
}

func knownTypeKeys() (out []string) {
	for k := range namespacedTypes {
		out = append(out, k)
	}

	for k := range globalTypes {
		out = append(out, k)
	}

	return
}

func getResourceInterface(client dynamic.Interface, k8sType, ns string) (dynamic.ResourceInterface, error) {
	if resource, ok := namespacedTypes[k8sType]; ok {
		return client.Resource(resource).Namespace(ns), nil
	}

	if resource, ok := globalTypes[k8sType]; ok {
		return client.Resource(resource), nil
	}

	return nil, fmt.Errorf("unknown type: %s", k8sType)
}

// BashCompletionFuncName gets the name of a bash completion func for the given
// type.
func BashCompletionFuncName(k8sType string) string {
	return fmt.Sprintf("__kf_name_%s", k8sType)
}

// BashCompletionFunc returns the bash completion function for a single type.
func BashCompletionFunc(k8sType string) string {
	return BashCompletionFuncName(k8sType) + `()
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
	return flags.SetAnnotation(name, cobra.BashCompCustom, []string{BashCompletionFuncName(k8sType)})
}

// MarkArgCompletionSupported returns completion annotations for a CobraCommand
func MarkArgCompletionSupported(cmd *cobra.Command, k8sType string) {
	if cmd == nil {
		return
	}

	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations[cobra.BashCompCustom] = BashCompletionFuncName(k8sType)
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
func AddBashCompletion(rootCommand *cobra.Command) {
	out := &bytes.Buffer{}

	for _, k8sType := range knownTypeKeys() {
		fmt.Fprintln(out, BashCompletionFunc(k8sType))
	}

	fmt.Fprintln(out)

	fmt.Fprintln(out, "__kf_custom_func() {")
	fmt.Fprintln(out, "    case ${last_command} in")

	for commandName, completionFunc := range customCompletions(rootCommand) {
		fmt.Fprintf(out, "    %s)\n", commandName)
		fmt.Fprintf(out, "        %s\n", completionFunc)
		fmt.Fprintln(out, "        return")
		fmt.Fprintln(out, "        ;;")
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, "    *)")
	fmt.Fprintln(out, "        ;;")
	fmt.Fprintln(out, "    esac")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	rootCommand.BashCompletionFunction = out.String()
}

// NewNamesCommand generates a command to get the names of various types
func NewNamesCommand(p *config.KfParams, client dynamic.Interface) *cobra.Command {
	return &cobra.Command{
		Hidden: true,

		Use:       "names TYPE",
		Short:     "Generates names for autocomplete",
		Example:   `kf names apps`,
		Long:      `This command is used by shell autocompletion`,
		ValidArgs: knownTypeKeys(),
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cobra.OnlyValidArgs(cmd, args); err != nil {
				return err
			}

			k8sType := args[0]

			client, err := getResourceInterface(client, k8sType, p.Namespace)
			if err != nil {
				return err
			}

			ul, err := client.List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			var names []string
			for _, li := range ul.Items {
				names = append(names, li.GetName())
			}

			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(names, " "))

			return nil
		},
	}
}
