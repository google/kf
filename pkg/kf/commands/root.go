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
	"runtime"
	"strings"

	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/doctor"
	"github.com/google/kf/pkg/kf/commands/group"
	"github.com/google/kf/pkg/kf/commands/install"
	pkgdoctor "github.com/google/kf/pkg/kf/doctor"
	"github.com/google/kf/pkg/kf/istio"
	templates "github.com/google/kf/third_party/kubectl-templates"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	p := &config.KfParams{}

	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience",
		Long: templates.LongDesc(`
			Kf is a MicroPaaS for Kubernetes with a Cloud Foundry style developer
			expeience.

			Kf aims to be fully compatible with Cloud Foundry applications and
			lifecycle. It supports logs, buildpacks, app manifests, routing, service
			brokers, and injected services.

			At the same time, it aims to improve the operational experience by
			supporting git-ops, self-healing infrastructure, containers, a service
			mesh, autoscaling, scale-to-zero, improved quota management and does it
			all on Kubernetes using industry-standard OSS tools including Knative,
			Istio, and Tekton.
			`),
		DisableAutoGenTag: false,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			loadedConfig, err := config.Load(p.Config, p)
			if err != nil {
				return err
			}

			return mergo.Map(p, loadedConfig)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&p.Config, "config", "", "Config file (default is $HOME/.kf)")
	rootCmd.PersistentFlags().StringVar(&p.KubeCfgFile, "kubeconfig", "", "Kubectl config file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "", "Kubernetes namespace to target")
	if err := completion.MarkFlagCompletionSupported(rootCmd.PersistentFlags(), "namespace", "spaces"); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().BoolVar(&p.LogHTTP, "log-http", false, "Log HTTP requests to stderr")

	rootCmd = group.AddCommandGroups(rootCmd, group.CommandGroups{
		{
			Name: "App Management",
			Commands: []*cobra.Command{
				InjectPush(p),
				InjectDelete(p),
				InjectApps(p),
				InjectGetApp(p),
				InjectStart(p),
				InjectStop(p),
				InjectRestart(p),
				InjectRestage(p),
				InjectScale(p),
				InjectLogs(p),
				InjectProxy(p),
			},
		},
		{
			Name: "Environment Variables",
			Commands: []*cobra.Command{
				InjectEnv(p),
				InjectSetEnv(p),
				InjectUnsetEnv(p),
			},
		},
		{
			Name: "Buildpacks",
			Commands: []*cobra.Command{
				InjectBuildpacks(p),
				InjectStacks(p),
			},
		},
		{
			Name: "Routing",
			Commands: []*cobra.Command{
				InjectRoutes(p),
				InjectCreateRoute(p),
				InjectDeleteRoute(p),
				InjectMapRoute(p),
				InjectUnmapRoute(p),
				InjectProxyRoute(p),
			},
		},
		{
			Name: "Quotas",
			Commands: []*cobra.Command{
				InjectGetQuota(p),
				InjectUpdateQuota(p),
				InjectDeleteQuota(p),
			},
		},
		{
			Name: "Services",
			Commands: []*cobra.Command{
				InjectCreateService(p),
				InjectDeleteService(p),
				InjectGetService(p),
				InjectListServices(p),
				InjectMarketplace(p),
			},
		},
		{
			Name: "Service Bindings",
			Commands: []*cobra.Command{
				InjectBindingService(p),
				InjectListBindings(p),
				InjectUnbindService(p),
				InjectVcapServices(p),
			},
		},
		{
			Name: "Service Brokers",
			Commands: []*cobra.Command{
				InjectCreateServiceBroker(p),
				InjectDeleteServiceBroker(p),
			},
		},
		{
			Name: "Spaces",
			Commands: []*cobra.Command{
				InjectSpaces(p),
				InjectSpace(p),
				InjectCreateSpace(p),
				InjectDeleteSpace(p),
				InjectConfigSpace(p),
			},
		},
		{
			Name: "Builds",
			Commands: []*cobra.Command{
				InjectBuilds(p),
				InjectBuildLogs(p),
				InjectBuild(p),
			},
		},
		{
			Name: "Other Commands",
			Commands: []*cobra.Command{
				// DoctorTests are run in the order they're defined in this list.
				// Tests will stop as soon as one of these top-level tests fails so they
				// should be ordered in a logical way e.g. testing apps should come after
				// testing the cluster because if the cluster isn't working then all the
				// app tests will fail.
				doctor.NewDoctorCommand(p, []doctor.DoctorTest{
					{Name: "cluster", Test: pkgdoctor.NewClusterDiagnostic(config.GetKubernetes(p))},
					{Name: "buildpacks", Test: InjectBuildpacksClient(p)},
					{Name: "istio", Test: istio.NewIstioClient(config.GetKubernetes(p))},
				}),

				completionCommand(rootCmd),
				install.NewInstallCommand(),
				NewTargetCommand(p),
				NewVersionCommand(Version, runtime.GOOS),
				NewDebugCommand(p, config.GetKubernetes(p)),
				InjectNamesCommand(p),
			},
		},
	})

	completion.AddBashCompletion(rootCmd)

	// We don't want the AutoGenTag as it makes the doc generation
	// non-deterministic. We would rather allow the CI to ensure the docs were
	// regenerated for each commit.
	rootCmd.DisableAutoGenTag = true

	rootCmd = templates.NormalizeAll(rootCmd)

	return rootCmd
}

func completionCommand(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion bash|zsh",
		Short: "Generate auto-completion files for kf commands",
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
