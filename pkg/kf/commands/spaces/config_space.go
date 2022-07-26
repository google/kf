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
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/algorithms"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmp"
	k8syaml "sigs.k8s.io/yaml"
)

// NewConfigSpaceCommand creates a command that can set facets of a space.
func NewConfigSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure-space [subcommand]",
		Aliases: []string{"config-space"},
		Short:   "Set configuration for a Space.",
		Long: `The configure-space sub-command allows operators to configure
		individual fields on a Space.

		In Kf, most configuration can be overridden at the Space level.

		NOTE: The Space is reconciled every time changes are made using this command.
		If you want to configure Spaces in automation it's better to use kubectl.
		`,
		SilenceUsage: true,
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
		newAppendDomainMutator(),
		newSetDefaultDomainMutator(),
		newRemoveDomainMutator(),
		newBuildServiceAccountMutator(),
		newSetAppIngressPolicyMutator(),
		newSetAppEgressPolicyMutator(),
		newSetBuildIngressPolicyMutator(),
		newSetBuildEgressPolicyMutator(),
		newSetNodeSelectorMutator(),
		newUnsetNodeSelectorMutator(),
	}

	for _, sm := range subcommands {
		cmd.AddCommand(sm.toCommand(p, client))
	}

	accessors := []spaceAccessor{
		newGetContainerRegistryAccessor(),
		newGetExecutionEnvAccessor(),
		newGetBuildpackEnvAccessor(),
		newGetDomainsAccessor(),
		newGetBuildServiceAccountAccessor(),
		newGetNodeSelectorAccessor(),
	}

	for _, sa := range accessors {
		cmd.AddCommand(sa.toCommand(p, client))
	}

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
	fmt.Fprintf(buffer, "# Configure the Space \"my-space\"\n")
	fmt.Fprintf(buffer, "kf configure-space %s my-space %s\n", sm.Name, joinedArgs)
	fmt.Fprintf(buffer, "# Configure the targeted Space\n")
	fmt.Fprintf(buffer, "kf configure-space %s %s\n", sm.Name, joinedArgs)
	return buffer.String()
}

func (sm spaceMutator) toCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s [SPACE_NAME] %s", sm.Name, strings.Join(sm.Args, " ")),
		Short:             sm.Short,
		Long:              sm.Short,
		Args:              cobra.RangeArgs(len(sm.Args), 1+len(sm.Args)),
		Example:           sm.exampleCommands(),
		ValidArgsFunction: completion.SpaceCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var spaceName string
			if len(args) <= len(sm.Args) {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
				spaceName = p.Space
			} else {
				spaceName = args[0]
				args = args[1:]
			}

			mutator, err := sm.Init(args)
			if err != nil {
				return err
			}

			diffPrintingMutator := DiffWrapper(cmd.OutOrStdout(), mutator)
			_, err = client.Transform(cmd.Context(), spaceName, diffPrintingMutator)
			if err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.OutOrStdout(), "configuring Space", func() error {
				_, err := client.WaitForConditionReadyTrue(context.Background(), spaceName, 1*time.Second)
				return err
			})
		},
	}
	async.Add(cmd)

	return cmd
}

func newSetContainerRegistryMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-container-registry",
		Short:       "Set the container registry used for Builds.",
		Args:        []string{"REGISTRY"},
		ExampleArgs: []string{"gcr.io/my-project"},
		Init: func(args []string) (spaces.Mutator, error) {
			registry := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildConfig.ContainerRegistry = registry

				return nil
			}, nil
		},
	}
}

func newSetEnvMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-env",
		Short:       "Set a Space wide environment variable for all Apps.",
		Args:        []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		ExampleArgs: []string{"ENVIRONMENT", "production"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]
			value := args[1]

			return func(space *v1alpha1.Space) error {
				tmp := envutil.RemoveEnvVars([]string{name}, space.Spec.RuntimeConfig.Env)
				space.Spec.RuntimeConfig.Env = append(tmp, corev1.EnvVar{Name: name, Value: value})

				return nil
			}, nil
		},
	}
}

func newUnsetEnvMutator() spaceMutator {
	return spaceMutator{
		Name:        "unset-env",
		Short:       "Unset a Space wide environment variable for all Apps.",
		Args:        []string{"ENV_VAR_NAME"},
		ExampleArgs: []string{"ENVIRONMENT"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.RuntimeConfig.Env = envutil.RemoveEnvVars([]string{name}, space.Spec.RuntimeConfig.Env)

				return nil
			}, nil
		},
	}
}

func newSetNodeSelectorMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-nodeselector",
		Short:       "Set a Space wide node selector for all Apps.",
		Args:        []string{"NS_VAR_NAME", "NS_VAR_VALUE"},
		ExampleArgs: []string{"DiskType", "ssd"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]
			value := args[1]

			return func(space *v1alpha1.Space) error {
				if space.Spec.RuntimeConfig.NodeSelector == nil {
					space.Spec.RuntimeConfig.NodeSelector = make(map[string]string)
				}
				space.Spec.RuntimeConfig.NodeSelector[name] = value
				return nil
			}, nil
		},
	}
}

func newUnsetNodeSelectorMutator() spaceMutator {
	return spaceMutator{
		Name:        "unset-nodeselector",
		Short:       "Unset a Space wide node selector for all Apps.",
		Args:        []string{"NS_VAR_NAME"},
		ExampleArgs: []string{"DiskType"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				delete(space.Spec.RuntimeConfig.NodeSelector, name)
				return nil
			}, nil
		},
	}
}

func newSetBuildpackEnvMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-buildpack-env",
		Short:       "Set an environment variable for Builds in a Space.",
		Args:        []string{"ENV_VAR_NAME", "ENV_VAR_VALUE"},
		ExampleArgs: []string{"JDK_VERSION", "11"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]
			value := args[1]

			return func(space *v1alpha1.Space) error {
				tmp := envutil.RemoveEnvVars([]string{name}, space.Spec.BuildConfig.Env)
				space.Spec.BuildConfig.Env = append(tmp, corev1.EnvVar{Name: name, Value: value})

				return nil
			}, nil
		},
	}
}

func newUnsetBuildpackEnvMutator() spaceMutator {
	return spaceMutator{
		Name:        "unset-buildpack-env",
		Short:       "Unset an environment variable for Builds in a Space.",
		Args:        []string{"ENV_VAR_NAME"},
		ExampleArgs: []string{"JDK_VERSION"},
		Init: func(args []string) (spaces.Mutator, error) {
			name := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildConfig.Env = envutil.RemoveEnvVars([]string{name}, space.Spec.BuildConfig.Env)

				return nil
			}, nil
		},
	}
}

func newAppendDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "append-domain",
		Short:       "Append a domain for a Space.",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.Domains = append(
					space.Spec.NetworkConfig.Domains,
					v1alpha1.SpaceDomain{Domain: domain},
				)

				return nil
			}, nil
		},
	}
}

func newSetAppIngressPolicyMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-app-ingress-policy",
		Short:       "Set the ingress policy for Apps in the Space.",
		Args:        []string{"POLICY"},
		ExampleArgs: []string{"DenyAll"},
		Init: func(args []string) (spaces.Mutator, error) {
			policy := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.AppNetworkPolicy.Ingress = policy
				return nil
			}, nil
		},
	}
}

func newSetAppEgressPolicyMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-app-egress-policy",
		Short:       "Set the egress policy for Apps in the Space.",
		Args:        []string{"POLICY"},
		ExampleArgs: []string{"DenyAll"},
		Init: func(args []string) (spaces.Mutator, error) {
			policy := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.AppNetworkPolicy.Egress = policy
				return nil
			}, nil
		},
	}
}

func newSetBuildIngressPolicyMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-build-ingress-policy",
		Short:       "Set the ingress policy for Builds in the Space.",
		Args:        []string{"POLICY"},
		ExampleArgs: []string{"DenyAll"},
		Init: func(args []string) (spaces.Mutator, error) {
			policy := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.BuildNetworkPolicy.Ingress = policy
				return nil
			}, nil
		},
	}
}

func newSetBuildEgressPolicyMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-build-egress-policy",
		Short:       "Set the egress policy for Builds in the Space.",
		Args:        []string{"POLICY"},
		ExampleArgs: []string{"DenyAll"},
		Init: func(args []string) (spaces.Mutator, error) {
			policy := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.BuildNetworkPolicy.Egress = policy
				return nil
			}, nil
		},
	}
}

func newSetDefaultDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "set-default-domain",
		Short:       "Set or create a default domain for a Space.",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {

				var tmp []v1alpha1.SpaceDomain
				tmp = append(tmp, v1alpha1.SpaceDomain{
					Domain: domain,
				})

				tmp = append(tmp, space.Spec.NetworkConfig.Domains...)
				space.Spec.NetworkConfig.Domains = v1alpha1.StableDeduplicateSpaceDomainList(tmp)
				return nil
			}, nil
		},
	}
}

func newRemoveDomainMutator() spaceMutator {
	return spaceMutator{
		Name:        "remove-domain",
		Short:       "Remove a domain from a Space.",
		Args:        []string{"DOMAIN"},
		ExampleArgs: []string{"myspace.mycompany.com"},
		Init: func(args []string) (spaces.Mutator, error) {
			domain := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.NetworkConfig.Domains = []v1alpha1.SpaceDomain(algorithms.Delete(
					v1alpha1.SpaceDomains(space.Spec.NetworkConfig.Domains),
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
		Short:       "Set the service account to use when building containers.",
		Args:        []string{"SERVICE_ACCOUNT"},
		ExampleArgs: []string{"myserviceaccount"},
		Init: func(args []string) (spaces.Mutator, error) {
			serviceAccount := args[0]

			return func(space *v1alpha1.Space) error {
				space.Spec.BuildConfig.ServiceAccount = serviceAccount
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
	fmt.Fprintf(buffer, "# Configure the Space \"my-space\"\n")
	fmt.Fprintf(buffer, "kf configure-space %s my-space\n", sm.Name)
	fmt.Fprintf(buffer, "# Configure the targeted Space\n")
	fmt.Fprintf(buffer, "kf configure-space %s\n", sm.Name)
	return buffer.String()
}

func (sm spaceAccessor) toCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s [SPACE_NAME]", sm.Name),
		Short:             sm.Short,
		Long:              sm.Short,
		Example:           sm.exampleCommands(),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.SpaceCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var spaceName string
			if len(args) == 0 {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
				spaceName = p.Space
			} else {
				spaceName = args[0]
			}

			space, err := client.Get(cmd.Context(), spaceName)
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

	return cmd
}

func newGetContainerRegistryAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-container-registry",
		Short: "Get the container registry used for Builds.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildConfig.ContainerRegistry
		},
	}
}

func newGetExecutionEnvAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-execution-env",
		Short: "Get the Space wide App environment variables.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.RuntimeConfig.Env
		},
	}
}

func newGetBuildpackEnvAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-buildpack-env",
		Short: "Get the environment variables for Builds in a Space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildConfig.Env
		},
	}
}

func newGetDomainsAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-domains",
		Short: "Get domains associated with the Space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.NetworkConfig.Domains
		},
	}
}

func newGetBuildServiceAccountAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-build-service-account",
		Short: "Get the service account that is used when building containers in the Space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.BuildConfig.ServiceAccount
		},
	}
}

func newGetNodeSelectorAccessor() spaceAccessor {
	return spaceAccessor{
		Name:  "get-nodeselector",
		Short: "Get the node selector associated with the Space.",
		Accessor: func(space *v1alpha1.Space) interface{} {
			return space.Spec.RuntimeConfig.NodeSelector
		},
	}
}

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator spaces.Mutator) spaces.Mutator {
	return func(mutable *v1alpha1.Space) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two v1alpha1.Spaces and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1alpha1.Space) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "Space Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}
