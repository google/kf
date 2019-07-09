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

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/doctor"
	pkgdoctor "github.com/google/kf/pkg/kf/doctor"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	p := &config.KfParams{}

	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "kf is like cf for Knative",
		Long: templates.LongDesc(`
      kf is like cf for Knative
      `),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			loadedConfig, err := config.Load(p.Config, p)
			if err != nil {
				return err
			}

			return mergo.Map(p, loadedConfig)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&p.Config, "config", "", "config file (default is $HOME/.kf)")
	rootCmd.PersistentFlags().StringVar(&p.KubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "", "kubernetes namespace")

	groups := templates.CommandGroups{
		{
			Message: "App Management",
			Commands: []*cobra.Command{
				InjectPush(p),
				InjectDelete(p),
				InjectApps(p),
				InjectProxy(p),
				InjectLogs(p),
			},
		},
		{
			Message: "Environment Variables",
			Commands: []*cobra.Command{
				InjectEnv(p),
				InjectSetEnv(p),
				InjectUnsetEnv(p),
			},
		},
		{
			Message: "Buildpacks",
			Commands: []*cobra.Command{
				InjectBuildpacks(p),
				InjectStacks(p),
			},
		},
		{
			Message: "Routing",
			Commands: []*cobra.Command{
				InjectRoutes(p),
				InjectCreateRoute(p),
				InjectDeleteRoute(p),
				InjectMapRoute(p),
				InjectUnmapRoute(p),
			},
		},
		{
			Message: "Quotas",
			Commands: []*cobra.Command{
				InjectQuotas(p),
				InjectGetQuota(p),
				InjectCreateQuota(p),
				InjectUpdateQuota(p),
				InjectDeleteQuota(p),
			},
		},
		{
			Message: "Services",
			Commands: []*cobra.Command{
				InjectCreateService(p),
				InjectDeleteService(p),
				InjectGetService(p),
				InjectListServices(p),
				InjectMarketplace(p),
			},
		},
		{
			Message: "Service Bindings",
			Commands: []*cobra.Command{
				InjectBindingService(p),
				InjectListBindings(p),
				InjectUnbindService(p),
				InjectVcapServices(p),
			},
		},
		{
			Message: "Spaces",
			Commands: []*cobra.Command{
				InjectSpaces(p),
				InjectSpace(p),
				InjectCreateSpace(p),
				InjectDeleteSpace(p),
				InjectConfigSpace(p),
			},
		},
		{
			Message: "Other Commands",
			Commands: []*cobra.Command{
				// DoctorTests are run in the order they're defined in this list.
				// Tests will stop as soon as one of these top-level tests fails so they
				// should be ordered in a logical way e.g. testing apps should come after
				// testing the cluster because if the cluster isn't working then all the
				// app tests will fail.
				doctor.NewDoctorCommand(p, []doctor.DoctorTest{
					{Name: "cluster", Test: pkgdoctor.NewClusterDiagnostic(config.GetKubernetes(p))},
					{Name: "buildpacks", Test: InjectBuildpacksClient(p)},
				}),

				completionCommand(rootCmd),
				NewTargetCommand(p),
			},
		},
	}

	// This will add the rest to a group under "Other Commands".
	groups.Add(rootCmd)
	templates.ActsAsRootCommand(rootCmd, nil, groups...)

	return rootCmd
}

func completionCommand(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use: "completion bash|zsh",
		Example: `
  eval "$(kf completion bash)"
  eval "$(kf completion zsh)"
		`,
		Long: `completion is used to create set up bash/zsh auto-completion for kf commands.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch shell := strings.ToLower(args[0]); shell {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			default:
				return fmt.Errorf("unknown shell %q. Only bash and zsh are supported", shell)
			}
		},
	}
}
