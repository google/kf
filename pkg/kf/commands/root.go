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
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/commands/doctor"
	"github.com/google/kf/v2/pkg/kf/commands/group"
	"github.com/google/kf/v2/pkg/kf/commands/stubs"
	pkgdoctor "github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/third_party/k8s.io/kubectl/pkg/util/templates"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kninjection "knative.dev/pkg/injection"
	"knative.dev/pkg/logging"

	// The following imports provide authentication helpers for GCP and OIDC.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var k8sClientFactory = config.GetKubernetes

const (
	docsURL    = "https://cloud.google.com/migrate/kf"
	versionURL = "https://cloud.google.com/migrate/kf/docs/downloads"
)

// NewKfCommand creates the root Kf command suitable for using in a CLI context.
func NewKfCommand() *cobra.Command {
	cmd := NewRawKfCommand()
	return templates.NormalizeAll(cmd)
}

// NewRawKfCommand returns the root Kf command without Kubernetes style
// formatting applied to the docs which permanently modifies them by changing
// whitespace and running them through a custom markdown to text renderer.
func NewRawKfCommand() *cobra.Command {
	p := &config.KfParams{}

	var postRunOnce sync.Once
	var rootCmd = &cobra.Command{
		Use:                   "kf",
		Short:                 "Kf CLI",
		Long:                  fmt.Sprintf("Kf CLI Version: %s Documentation: %s .", Version, docsURL),
		DisableAutoGenTag:     false,
		TraverseChildrenHooks: true,
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			loadedConfig, err := config.Load(p.Config, p)
			if err != nil {
				return err
			}

			if err := mergo.Map(p, loadedConfig); err != nil {
				return err
			}

			ctx := kninjection.WithNamespaceScope(cmd.Context(), p.Space)
			ctx = config.SetupInjection(ctx, p)
			ctx = configlogging.SetupLogger(ctx, cmd.ErrOrStderr())
			cmd.SetContext(ctx)

			_, skipVersion := cmd.Annotations[config.SkipVersionCheckAnnotation]
			if !skipVersion {
				ns := getKfNamespace(ctx, k8sClientFactory(p))
				if ns == nil {
					return nil
				}

				checkVersion(ctx, ns)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			// NOTE: this hook gets executed by Cobra and also main.go if the command
			// fails because Cobra won't execute it if the main command returns an
			// error. The second args param is unusable if Cobra didn't call us so
			// it's assigned to _.
			postRunOnce.Do(func() {
				utils.PrintNextActions(cmd.ErrOrStderr())
			})
		},
	}

	rootCmd.PersistentFlags().StringVar(&p.Config, "config", "", "Path to the Kf config file to use for CLI requests.")
	rootCmd.PersistentFlags().StringVar(&p.KubeCfgFile, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")
	rootCmd.PersistentFlags().StringVar(&p.Space, "space", "", "Space to run the command against. This flag overrides the currently targeted Space.")
	rootCmd.PersistentFlags().StringVar(&p.Impersonate.UserName, "as", "", "Username to impersonate for the operation.")
	rootCmd.PersistentFlags().StringSliceVar(&p.Impersonate.Groups, "as-group", []string{}, "Group to impersonate for the operation. Include this flag multiple times to specify multiple groups.")

	rootCmd.RegisterFlagCompletionFunc("space", completion.SpaceCompletionFn(p))

	rootCmd.PersistentFlags().BoolVar(&p.LogHTTP, "log-http", false, "Log HTTP requests to standard error.")

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
			Name: "Auto Scale",
			Commands: []*cobra.Command{
				InjectEnableAutoscale(p),
				InjectDisableAutoscale(p),
				InjectCreateAutoscalingRule(p),
				InjectDeleteAutoscalingRules(p),
				InjectUpdateAutoscalingLimits(p),
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
				InjectDeleteOrphanedRoutes(p),
				InjectMapRoute(p),
				InjectUnmapRoute(p),
				InjectProxyRoute(p),
				InjectDomains(p),
			},
		},
		{
			Name: "Services",
			Commands: []*cobra.Command{
				InjectCreateService(p),
				InjectCreateUserProvidedService(p),
				InjectUpdateUserProvidedService(p),
				InjectDeleteService(p),
				InjectGetService(p),
				InjectListServices(p),
				InjectMarketplace(p),
			},
		},
		{
			Name: "Service Bindings",
			Commands: []*cobra.Command{
				InjectBindService(p),
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
				InjectTarget(p),
				InjectSpaces(p),
				InjectSpace(p),
				InjectCreateSpace(p),
				InjectDeleteSpace(p),
				InjectConfigSpace(p),
				InjectSpaceUsers(p),
				InjectSetSpaceRole(p),
				InjectUnsetSpaceRole(p),
			},
		},
		{
			Name: "Cluster",
			Commands: []*cobra.Command{
				InjectConfigCluster(p),
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
			Name: "Network Policies",
			Commands: []*cobra.Command{
				InjectNetworkPolicies(p),
				InjectDeleteNetworkPolicies(p),
				InjectDescribeNetworkPolicy(p),
			},
		},
		{
			Name: "Tasks",
			Commands: []*cobra.Command{
				InjectRunTask(p),
				InjectTerminateTask(p),
				InjectTasks(p),
			},
		},
		{
			Name: "Jobs",
			Commands: []*cobra.Command{
				InjectCreateJob(p),
				InjectRunJob(p),
				InjectScheduleJob(p),
				InjectListJobs(p),
				InjectListJobSchedules(p),
				InjectJobHistory(p),
				InjectDeleteJob(p),
				InjectDeleteJobSchedule(p),
			},
		},
		utils.PreviewCommandGroup(
			"Route Services",
			InjectBindRouteService(p),
			InjectUnbindRouteService(p),
		),
		utils.ExperimentalCommandGroup(
			"Buildpacks",
			InjectWrapV2Buildpack(p),
		),
		{
			Name: "Other Commands",
			Commands: []*cobra.Command{
				// DoctorTests are run in the order they're defined in this list.
				// Tests will stop as soon as one of these top-level tests fails so they
				// should be ordered in a logical way e.g. testing apps should come after
				// testing the cluster because if the cluster isn't working then all the
				// app tests will fail.
				doctor.NewDoctorCommand(
					p,
					[]doctor.DoctorTest{
						{Name: "cluster", Test: pkgdoctor.NewClusterDiagnostic(config.GetKubernetes(p))},
						{Name: "istio", Test: pkgdoctor.NewIstioDiagnostic(config.GetKubernetes(p))},
						{Name: "operator", Test: pkgdoctor.NewOperatorDiagnostic(config.GetKubernetes(p))},
					},
					troubleshooter.CustomResourceComponents(),
					troubleshooter.NewTroubleshootingCloser(p),
				),

				NewVersionCommand(Version, runtime.GOOS),
				NewDebugCommand(p, config.GetKubernetes(p)),
				stubs.NewLoginCommand(),
				stubs.NewApiCommand(),
				stubs.NewAuthCommand(),
				stubs.NewLogoutCommand(),
				NewThirdPartyLicensesCommand(),
				NewAboutCommand(),
				InjectSSH(p),
				InjectDependencyCommand(p),
			},
		},
	})

	// We don't want the AutoGenTag as it makes the doc generation
	// non-deterministic. We would rather allow the CI to ensure the docs were
	// regenerated for each commit.
	rootCmd.DisableAutoGenTag = true

	return rootCmd
}

func getKfNamespace(ctx context.Context, k kubernetes.Interface) *v1.Namespace {
	ns, err := k.CoreV1().Namespaces().Get(ctx, v1alpha1.KfNamespace, metav1.GetOptions{})
	if err != nil {
		logging.FromContext(ctx).Warnf(
			"Error getting %s namespace for CLI warnings: %s",
			v1alpha1.KfNamespace,
			err,
		)
		return nil
	}
	return ns
}

// Warn user if client and server semver versions are different
func checkVersion(ctx context.Context, ns *v1.Namespace) {
	clientVersion := Version
	serverVersion := ns.Labels[v1alpha1.VersionLabel]

	if clientVersion != serverVersion {
		logger := logging.FromContext(ctx)
		logger.Warnf(
			"Client version %s does not match server version %s",
			clientVersion,
			serverVersion,
		)

		logger.Warnf("Visit %s to download a matching version", versionURL)
	}
}
