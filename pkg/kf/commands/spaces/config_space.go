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
	"bytes"
	"fmt"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/quotas"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8syaml "sigs.k8s.io/yaml"
)

// NewConfigSpaceCommand creates a command that can set facets of a space.
func NewConfigSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure-space [subcommand]",
		Aliases: []string{"config-space"},
		Short:   "Set configuration for a space",
		Long: `The configure-space sub-command allows operators to configure
		individual fields on a space.

		In Kf, almost all configuration is at the space level as opposed to being
		globally set on the cluster.

		NOTE: The space is queued for reconciliation every time changes are made
		via this command. If you want to configure spaces in automation it's better
		to use kubectl.
		`,
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
		newAppendDomainMutator(),
		newSetDefaultDomainMutator(),
		newRemoveDomainMutator(),
		newBuildServiceAccountMutator(),
	}

	for _, sm := range subcommands {
		cmd.AddCommand(sm.ToCommand(p, client))
	}

	accessors := []spaceAccessor{
		newGetContainerRegistryAccessor(),
		newGetBuildpackBuilderAccessor(),
		newGetExecutionEnvAccessor(),
		newGetBuildpackEnvAccessor(),
		newGetDomainsAccessor(),
		newGetBuildServiceAccountAccessor(),
	}

	for _, sa := range accessors {
		cmd.AddCommand(sa.ToCommand(p, client))
	}

	cmd.AddCommand(
		quotas.NewGetQuotaCommand(p, client),
		quotas.NewUpdateQuotaCommand(p, client),
		quotas.NewDeleteQuotaCommand(p, client),
	)

	return cmd
}

type spaceMutator struct {
	Name        string
	Short       string
	Args        []string
	ExampleArgs []string
	Init        func(args []string) (spaces.Mutator, error)
}

func (sm spaceMutator) exampleCommands() string {
	joinedArgs := strings.Join(sm.ExampleArgs, " ")
	buffer := &bytes.Buffer{}
	fmt.Fprintln(buffer)
	fmt.Fprintf(buffer, "  # Configure the space \"my-space\"\n")
	fmt.Fprintf(buffer, "  kf configure-space %s my-space %s\n", sm.Name, joinedArgs)
	fmt.Fprintf(buffer, "  # Configure the targeted space\n")
	fmt.Fprintf(buffer, "  kf configure-space %s %s\n", sm.Name, joinedArgs)
	return buffer.String()
}

func (sm spaceMutator) ToCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [SPACE_NAME] %s", sm.Name, strings.Join(sm.Args, " ")),
		Short:   sm.Short,
		Long:    sm.Short,
		Args:    cobra.RangeArgs(len(sm.Args), 1+len(sm.Args)),
		Example: sm.exampleCommands(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var spaceName string
			if len(args) <= len(sm.Args) {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}
				spaceName = p.Namespace
			} else {
				spaceName = args[0]
				args = args[1:]
			}

			mutator, err := sm.Init(args)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			diffPrintingMutator := spaces.DiffWrapper(cmd.OutOrStdout(), mutator)
			_, err = client.Transform(spaceName, diffPrintingMutator)
			return err
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}

func newSetContainerRegistryMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-container-registry",
		Short:       "Set the container registry used for builds.",
		Args:        []string{"REGISTRY"},
		ExampleArgs: []string{"gcr.io/my-project"},
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
		Name:        "set-buildpack-builder",
		Short:       "Set the buildpack builder image.",
		Args:        []string{"BUILDER_IMAGE"},
		ExampleArgs: []string{"gcr.io/my-project/builder:latest"},
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
		Name:        "set-env",
		Short:       "Set a space-wide environment variable.",
		Args:        []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		ExampleArgs: []string{"ENVIRONMENT", "production"},
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
		Name:        "unset-env",
		Short:       "Unset a space-wide environment variable.",
		Args:        []string{"ENV_VAR_NAME"},
		ExampleArgs: []string{"ENVIRONMENT"},
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
		Name:        "set-buildpack-env",
		Short:       "Set an environment variable for buildpack builds in a space.",
		Args:        []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		ExampleArgs: []string{"JDK_VERSION", "11"},
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
		Name:        "unset-buildpack-env",
		Short:       "Unset an environment variable for buildpack builds in a space.",
		Args:        []string{"ENV_VAR_NAME"},
		ExampleArgs: []string{"JDK_VERSION"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildpackBuild.Env = envutil.RemoveEnvVars([]string{name}, space.Spec.BuildpackBuild.Env)

				return nil
			}, nil
		},
	}
}

func newAppendDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "append-domain",
		Short:       "Append a domain for a space",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.Execution.Domains = append(
					space.Spec.Execution.Domains,
					v1alpha1.SpaceDomain{Domain: domain},
				)

				return nil
			}, nil
		},
	}
}

func newSetDefaultDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-default-domain",
		Short:       "Set a default domain for a space",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {
				var found bool
				for i, d := range space.Spec.Execution.Domains {
					if d.Domain != domain {
						space.Spec.Execution.Domains[i].Default = false
						continue
					}
					found = true
					space.Spec.Execution.Domains[i].Default = true
				}

				if !found {
					return fmt.Errorf("failed to find domain %s", domain)
				}
				return nil
			}, nil
		},
	}
}

func newRemoveDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "remove-domain",
		Short:       "Remove a domain from a space",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.Execution.Domains = []v1alpha1.SpaceDomain(algorithms.Delete(
					v1alpha1.SpaceDomains(space.Spec.Execution.Domains),
					v1alpha1.SpaceDomains{{Domain: domain}},
				).(v1alpha1.SpaceDomains))

				return nil
			}, nil
		},
	}
}

func newBuildServiceAccountMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-build-service-account",
		Short:       "Set the service account to use when building containers",
		Args:        []string{"SERVICE_ACCOUNT"},
		ExampleArgs: []string{"myserviceaccount"},
		Init: func(args []string) (spaces.Mutator, error) {
			serviceAccount := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.Security.BuildServiceAccount = serviceAccount
				return nil
			}, nil
		},
	}
}

type spaceAccessor struct {
	Name     string
	Short    string
	Accessor func(space *v1alpha1.Space) interface{}
}

func (sm spaceAccessor) exampleCommands() string {
	buffer := &bytes.Buffer{}
	fmt.Fprintln(buffer)
	fmt.Fprintf(buffer, "  # Configure the space \"my-space\"\n")
	fmt.Fprintf(buffer, "  kf configure-space %s my-space\n", sm.Name)
	fmt.Fprintf(buffer, "  # Configure the targeted space\n")
	fmt.Fprintf(buffer, "  kf configure-space %s\n", sm.Name)
	return buffer.String()
}

func (sm spaceAccessor) ToCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [SPACE_NAME]", sm.Name),
		Short:   sm.Short,
		Long:    sm.Short,
		Example: sm.exampleCommands(),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var spaceName string
			if len(args) == 0 {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}
				spaceName = p.Namespace
			} else {
				spaceName = args[0]
			}

			cmd.SilenceUsage = true

			space, err := client.Get(spaceName)
			if err != nil {
				return err
			}

			out := sm.Accessor(space)

			// NOTE: use the K8s YAML marshal function because it works with builtin
			// k8s types by marshaling using the JSON tags then converting to YAML
			// as opposed to just using YAML tags natively.
			m, err := k8syaml.Marshal(out)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%#v", out)
				return fmt.Errorf("couldn't convert value to YAML: %s", err)
			}

			fmt.Fprint(cmd.OutOrStdout(), string(m))
			return nil
		},
	}

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}

func newGetContainerRegistryAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-container-registry",
		Short: "Get the container registry used for builds.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildpackBuild.ContainerRegistry
		},
	}
}

func newGetBuildpackBuilderAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-buildpack-builder",
		Short: "Get the buildpack builder used for builds.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildpackBuild.BuilderImage
		},
	}
}

func newGetExecutionEnvAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-execution-env",
		Short: "Get the space-wide environment variables.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.Execution.Env
		},
	}
}

func newGetBuildpackEnvAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-buildpack-env",
		Short: "Get the environment variables for buildpack builds in a space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildpackBuild.Env
		},
	}
}

func newGetDomainsAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-domains",
		Short: "Get domains associated with the space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.Execution.Domains
		},
	}
}

func newGetBuildServiceAccountAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-build-service-account",
		Short: "Get the service account that is used when building containers in the space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.Security.BuildServiceAccount
		},
	}
}
