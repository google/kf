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

package cluster

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/configmaps"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmp"
	k8syaml "sigs.k8s.io/yaml"
)

// NewConfigClusterCommand creates a command that can set facets of a Kf cluster.
func NewConfigClusterCommand(p *config.KfParams, client configmaps.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure-cluster [subcommand]",
		Aliases: []string{"config-cluster"},
		Short:   "Set configuration for a cluster.",
		Long: `
		The configure-cluster sub-command allows operators to configure
		individual fields on a cluster.
		`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	subcommands := []configMutator{
		newSetFeatureFlagMutator(),
		newUnsetFeatureFlagMutator(),
	}

	for _, sm := range subcommands {
		cmd.AddCommand(sm.toCommand(p, client))
	}

	accessors := []configAccessor{
		newGetFeatureFlagsAccessor(),
	}

	for _, sa := range accessors {
		cmd.AddCommand(sa.toCommand(p, client))
	}

	return cmd
}

type configMutator struct {
	Name          string
	Short         string
	Args          []string
	ExampleArgs   []string
	ConfigMapName string
	Init          func(args []string) (configmaps.Mutator, error)
}

func (cm configMutator) exampleCommands() string {
	joinedArgs := strings.Join(cm.ExampleArgs, " ")
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "# Configure the cluster.\n")
	fmt.Fprintf(buffer, "kf configure-cluster %s %s\n", cm.Name, joinedArgs)
	return buffer.String()
}

func (cm configMutator) toCommand(p *config.KfParams, client configmaps.Client) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s %s", cm.Name, strings.Join(cm.Args, " ")),
		Short:             cm.Short,
		Long:              cm.Short,
		Args:              cobra.ExactArgs(len(cm.Args)),
		Example:           cm.exampleCommands(),
		ValidArgsFunction: completion.SpaceCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configMapName := cm.ConfigMapName
			mutator, err := cm.Init(args)
			if err != nil {
				return err
			}
			diffPrintingMutator := DiffWrapper(cmd.OutOrStdout(), mutator)

			_, err = client.Transform(cmd.Context(), v1alpha1.KfNamespace, configMapName, diffPrintingMutator)
			if err != nil {
				return fmt.Errorf("error configuring cluster: %s", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Successfully configured cluster.")
			return nil
		},
	}
	async.Add(cmd)

	return cmd
}

func newSetFeatureFlagMutator() configMutator {
	return configMutator{
		Name:          "set-feature-flag",
		Short:         "Set a feature flag on the cluster.",
		Args:          []string{"FEATURE-FLAG-NAME", "BOOL"},
		ExampleArgs:   []string{"enable_route_services", "true"},
		ConfigMapName: kfconfig.DefaultsConfigName,
		Init: func(args []string) (configmaps.Mutator, error) {
			ffName := args[0]
			setBool, err := strconv.ParseBool(args[1])
			if err != nil {
				return nil, err
			}

			return func(cm *v1.ConfigMap) error {
				defaultsConfig, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
				if err != nil {
					return err
				}

				defaultsConfig.FeatureFlags[ffName] = setBool
				err = defaultsConfig.PatchConfigMap(cm)
				if err != nil {
					return err
				}

				return nil
			}, nil
		},
	}
}

func newUnsetFeatureFlagMutator() configMutator {
	return configMutator{
		Name:          "unset-feature-flag",
		Short:         "Unset a feature flag on the cluster. Resets feature flag value to default.",
		Args:          []string{"FEATURE-FLAG-NAME"},
		ExampleArgs:   []string{"enable_route_service"},
		ConfigMapName: kfconfig.DefaultsConfigName,
		Init: func(args []string) (configmaps.Mutator, error) {
			ffName := args[0]

			return func(cm *v1.ConfigMap) error {
				defaultsConfig, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
				if err != nil {
					return err
				}

				delete(defaultsConfig.FeatureFlags, ffName)
				return nil
			}, nil
		},
	}
}

type configAccessor struct {
	Name          string
	Short         string
	ConfigMapName string
	Accessor      func(cm *v1.ConfigMap) (interface{}, error)
}

func (ca configAccessor) exampleCommands() string {
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "# Configure the cluster.\n")
	fmt.Fprintf(buffer, "kf configure-cluster %s\n", ca.Name)
	return buffer.String()
}

func (ca configAccessor) toCommand(p *config.KfParams, client configmaps.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s", ca.Name),
		Short:             ca.Short,
		Long:              ca.Short,
		Example:           ca.exampleCommands(),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.SpaceCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm, err := client.Get(cmd.Context(), v1alpha1.KfNamespace, ca.ConfigMapName)
			if err != nil {
				return err
			}

			out, err := ca.Accessor(cm)
			if err != nil {
				return fmt.Errorf("couldn't access config: %s", err)
			}

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

func newGetFeatureFlagsAccessor() configAccessor {
	return configAccessor{
		Name:          "get-feature-flags",
		Short:         "Get the values for feature flags set on the cluster.",
		ConfigMapName: kfconfig.DefaultsConfigName,
		Accessor: func(cm *v1.ConfigMap) (interface{}, error) {
			defaultsConfig, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
			if err != nil {
				return nil, err
			}
			return defaultsConfig.FeatureFlags, nil
		},
	}
}

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator configmaps.Mutator) configmaps.Mutator {
	return func(mutable *v1.ConfigMap) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two ConfigMaps and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1.ConfigMap) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "ConfigMap Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}
