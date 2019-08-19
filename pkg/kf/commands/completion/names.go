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
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// NewNamesCommand generates a command to get the names of various types
func NewNamesCommand(p *config.KfParams, client dynamic.Interface) *cobra.Command {
	return &cobra.Command{
		Hidden: true,

		Use:     "names TYPE",
		Short:   "Get a list of names in the cluster for the given type",
		Example: `kf names apps`,
		Long: `The names command gets a list of the objects and prints the names in
		alphabetical order.

		If the type is namespaced, the objects in the targeted space are printed.`,
		ValidArgs: KnownGenericTypes(),
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

			PrintNames(cmd.OutOrStdout(), ul)

			return nil
		},
	}
}

// PrintNames prints the names of objects in the given list in alphabetical
// order.
func PrintNames(w io.Writer, ul *unstructured.UnstructuredList) {
	var names []string
	for _, li := range ul.Items {
		names = append(names, li.GetName())
	}

	sort.Strings(names)

	fmt.Fprintln(w, strings.Join(names, " "))
}
