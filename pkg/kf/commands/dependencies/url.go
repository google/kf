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

package dependencies

import (
	"fmt"
	"strings"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

// newURLCommand creates a command that gets the download
// URL for a dependency
func newURLCommand(dependencies []dependency) *cobra.Command {
	knownDependencies := sets.NewString()
	for _, d := range dependencies {
		knownDependencies = knownDependencies.Union(d.names())
	}

	cmd := &cobra.Command{
		Hidden: true,
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		Use:          "url [DEPENDENCY]",
		Short:        "Get the download URL for a dependency",
		Long:         documentationOnly,
		ValidArgs:    knownDependencies.List(), // List() is sorted
		Args:         cobra.ExactValidArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			depName := args[0]

			for _, dep := range dependencies {
				if dep.names().Has(depName) {
					_, url, err := dep.ResolveAll()
					if err != nil {
						return err
					}

					fmt.Fprintln(cmd.OutOrStdout(), url)
					return nil
				}
			}

			return fmt.Errorf(
				"unknown dependency %q, known dependencies are: %s",
				depName,
				strings.Join(knownDependencies.List(), ","),
			)
		},
	}

	return cmd
}
