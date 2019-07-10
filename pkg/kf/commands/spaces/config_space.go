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
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/envutil"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmp"
)

// NewConfigSpaceCommand creates a command that can set facets of a space.
func NewConfigSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure-space [subcommand]",
		Aliases: []string{"config-space"},
		Short:   "Set configuration for a space",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	subcommands := []spaceMutator{
		newSetEnvMutator(),
		newUnsetEnvMutator(),
		newSetBuildpackEnvMutator(),
		newUnsetBuildpackEnvMutator(),
		newSetContainerRegistryMutator(),
		newSetBuildpackBuilderMutator(),
	}

	for _, sm := range subcommands {
		cmd.AddCommand(sm.ToCommand(client))
	}

	return cmd
}

type spaceMutator struct {
	Name  string
	Short string
	Args  []string
	Init  func(args []string) (spaces.Mutator, error)
}

func (sm spaceMutator) ToCommand(client spaces.Client) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s SPACE_NAME %s", sm.Name, strings.Join(sm.Args, " ")),
		Short: sm.Short,
		Long:  sm.Short,
		Args:  cobra.ExactArgs(1 + len(sm.Args)),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceName := args[0]

			mutator, err := sm.Init(args[1:])
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			return client.Transform(spaceName, func(space *v1alpha1.Space) error {
				before := space.DeepCopy()

				if err := mutator(space); err != nil {
					return err
				}

				printDiff(cmd.OutOrStdout(), before, space)

				return nil
			})
		},
	}
}

func printDiff(w io.Writer, original, new *v1alpha1.Space) {
	diff, err := kmp.SafeDiff(original.Spec, new.Spec)
	if err != nil {
		fmt.Fprintf(w, "Couldn't format diff: %s\n", err.Error())
		return
	}

	if diff != "" {
		fmt.Fprintln(w, "Space Spec (-original +new):")
		fmt.Fprintln(w, diff)
	} else {
		fmt.Fprintln(w, "No changes")
	}
}

func newSetContainerRegistryMutator() spaceMutator {
	return spaceMutator{
		Name:  "set-container-registry",
		Short: "Set the container registry used for builds.",
		Args:  []string{"REGISTRY"},
		Init: func(args []string) (spaces.Mutator, error) {
			registry := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildpackBuild.ContainerRegistry = registry

				return nil
			}, nil
		},
	}
}

func newSetBuildpackBuilderMutator() spaceMutator {
	return spaceMutator{
		Name:  "set-buildpack-builder",
		Short: "Set the buildpack builder image.",
		Args:  []string{"BUILDER_IMAGE"},
		Init: func(args []string) (spaces.Mutator, error) {
			image := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildpackBuild.BuilderImage = image

				return nil
			}, nil
		},
	}
}

func newSetEnvMutator() spaceMutator {
	return spaceMutator{
		Name:  "set-env",
		Short: "Set a space-wide environment variable.",
		Args:  []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]
			value := args[1]

			return func(space *v1alpha1.Space) error {
				tmp := envutil.RemoveEnvVars([]string{name}, space.Spec.Execution.Env)
				space.Spec.Execution.Env = append(tmp, corev1.EnvVar{Name: name, Value: value})

				return nil
			}, nil
		},
	}
}

func newUnsetEnvMutator() spaceMutator {
	return spaceMutator{
		Name:  "unset-env",
		Short: "Unset a space-wide environment variable.",
		Args:  []string{"ENV_VAR_NAME"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.Execution.Env = envutil.RemoveEnvVars([]string{name}, space.Spec.Execution.Env)

				return nil
			}, nil
		},
	}
}

func newSetBuildpackEnvMutator() spaceMutator {
	return spaceMutator{
		Name:  "set-buildpack-env",
		Short: "Set an environment variable for buildpack builds in a space.",
		Args:  []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]
			value := args[1]

			return func(space *v1alpha1.Space) error {
				tmp := envutil.RemoveEnvVars([]string{name}, space.Spec.BuildpackBuild.Env)
				space.Spec.BuildpackBuild.Env = append(tmp, corev1.EnvVar{Name: name, Value: value})

				return nil
			}, nil
		},
	}
}

func newUnsetBuildpackEnvMutator() spaceMutator {
	return spaceMutator{
		Name:  "unset-buildpack-env",
		Short: "Unset an environment variable for buildpack builds in a space.",
		Args:  []string{"ENV_VAR_NAME"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildpackBuild.Env = envutil.RemoveEnvVars([]string{name}, space.Spec.BuildpackBuild.Env)

				return nil
			}, nil
		},
	}
}
