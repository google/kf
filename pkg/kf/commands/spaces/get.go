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

package spaces

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta/table"

	"github.com/spf13/cobra"
)

// NewGetSpaceCommand allows users to create spaces.
func NewGetSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space SPACE",
		Short: "Show space info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			space, err := client.Get(name)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			fmt.Fprintln(w, "# Metadata")
			fmt.Fprintf(w, "Name: %s\n", space.Name)
			fmt.Fprintf(w, "Age: %s\n", table.ConvertToHumanReadableDateType(space.CreationTimestamp))

			ready := ""
			reason := ""
			if cond := space.Status.GetCondition(v1alpha1.SpaceConditionReady); cond != nil {
				ready = fmt.Sprintf("%v", cond.Status)
				reason = cond.Reason
			}

			fmt.Fprintf(w, "Ready?: %s\n", ready)
			fmt.Fprintf(w, "Reason: %q\n", reason)

			fmt.Fprintln(w)
			fmt.Fprintln(w, "# Security")
			security := space.Spec.Security
			fmt.Fprintf(w, "Developers can read logs? %v\n", security.EnableDeveloperLogsAccess)

			fmt.Fprintln(w)
			fmt.Fprintln(w, "# Build")
			buildpackBuild := space.Spec.BuildpackBuild
			fmt.Fprintf(w, "Builder image: %q\n", buildpackBuild.BuilderImage)
			fmt.Fprintf(w, "Container registry: %q\n", buildpackBuild.ContainerRegistry)
			fmt.Fprintf(w, "Build environment: %v variable(s)\n", len(buildpackBuild.Env))
			printEnvGroup(w, buildpackBuild.Env)

			fmt.Fprintln(w)
			fmt.Fprintln(w, "# Execution")
			execution := space.Spec.Execution
			fmt.Fprintf(w, "Environment: %v variable(s)\n", len(execution.Env))
			printEnvGroup(w, execution.Env)

			return nil
		},
	}

	return cmd
}

func printEnvGroup(out io.Writer, envVars []corev1.EnvVar) {
	if len(envVars) == 0 {
		return
	}

	w := tabwriter.NewWriter(out, 8, 4, 1, ' ', tabwriter.StripEscape)
	defer w.Flush()

	fmt.Fprintln(w, "Variable Name\tAssigned Value")

	sort.Slice(envVars, func(i int, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})

	for _, env := range envVars {
		fmt.Fprintf(w, "%s\t%s", env.Name, env.Value)
		fmt.Fprintln(w)
	}
}
