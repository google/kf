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

//go:build wireinject
// +build wireinject

package commands

import (
	"github.com/google/go-containerregistry/pkg/v1/remote"
	kfv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/buildpacks"
	"github.com/google/kf/v2/pkg/kf/builds"
	capps "github.com/google/kf/v2/pkg/kf/commands/apps"
	"github.com/google/kf/v2/pkg/kf/commands/autoscaling"
	cbuildpacks "github.com/google/kf/v2/pkg/kf/commands/buildpacks"
	cbuilds "github.com/google/kf/v2/pkg/kf/commands/builds"
	ccluster "github.com/google/kf/v2/pkg/kf/commands/cluster"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	dependencies "github.com/google/kf/v2/pkg/kf/commands/dependencies"
	"github.com/google/kf/v2/pkg/kf/commands/exporttok8s"
	clogs "github.com/google/kf/v2/pkg/kf/commands/logs"
	cnetworkpolicies "github.com/google/kf/v2/pkg/kf/commands/networkpolicies"
	croutes "github.com/google/kf/v2/pkg/kf/commands/routes"
	servicebindingscmd "github.com/google/kf/v2/pkg/kf/commands/service-bindings"
	servicebrokerscmd "github.com/google/kf/v2/pkg/kf/commands/service-brokers"
	servicescmd "github.com/google/kf/v2/pkg/kf/commands/services"
	cspaces "github.com/google/kf/v2/pkg/kf/commands/spaces"
	ctasks "github.com/google/kf/v2/pkg/kf/commands/tasks"
	ctaskschedules "github.com/google/kf/v2/pkg/kf/commands/taskschedules"
	"github.com/google/kf/v2/pkg/kf/configmaps"
	kflogs "github.com/google/kf/v2/pkg/kf/logs"
	"github.com/google/kf/v2/pkg/kf/marketplace"
	"github.com/google/kf/v2/pkg/kf/routes"
	secrets "github.com/google/kf/v2/pkg/kf/secrets"
	clusterbrokerclient "github.com/google/kf/v2/pkg/kf/service-brokers/cluster"
	namespacedbrokerclient "github.com/google/kf/v2/pkg/kf/service-brokers/namespaced"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/google/kf/v2/pkg/kf/sourcepackages"
	"github.com/google/kf/v2/pkg/kf/spaces"
	tasks "github.com/google/kf/v2/pkg/kf/tasks"
	"github.com/google/wire"
	"github.com/spf13/cobra"
	k8sclient "k8s.io/client-go/kubernetes"
	corev1type "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func provideSrcImageBuilder() capps.SrcImageBuilder {
	return capps.SrcImageBuilderFunc(capps.DefaultSrcImageBuilder)
}

///////////////////
// App Commands //
/////////////////

var AppsSet = wire.NewSet(
	BuildsSet,
	provideAppsGetter,
	apps.NewClient,
	apps.NewPusher,
	kflogs.NewTailer,
)

func provideAppsGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.AppsGetter {
	return ki
}

func provideSourcePackagesGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.SourcePackagesGetter {
	return ki
}

func InjectPush(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewPushCommand,
		provideSrcImageBuilder,
		provideSecretsGetter,
		secrets.NewClient,
		provideServiceInstanceBindingsGetter,
		serviceinstancebindings.NewClient,
		sourcepackages.NewClient,
		provideSourcePackagesGetter,
		sourcepackages.NewPoster,
		AppsSet,
	)
	return nil
}

func InjectDelete(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewDeleteCommand)

	return nil
}

func InjectApps(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewAppsCommand)

	return nil
}

func InjectGetApp(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewGetAppCommand)

	return nil
}

func InjectScale(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewScaleCommand, AppsSet)
	return nil
}

func InjectCreateAutoscalingRule(p *config.KfParams) *cobra.Command {
	wire.Build(autoscaling.NewCreateAutoscalingRule, AppsSet)
	return nil
}

func InjectDeleteAutoscalingRules(p *config.KfParams) *cobra.Command {
	wire.Build(autoscaling.NewDeleteAutoscalingRules, AppsSet)
	return nil
}

func InjectUpdateAutoscalingLimits(p *config.KfParams) *cobra.Command {
	wire.Build(autoscaling.NewUpdateAutoscalingLimits, AppsSet)
	return nil
}

func InjectEnableAutoscale(p *config.KfParams) *cobra.Command {
	wire.Build(autoscaling.NewEnableAutoscaling, AppsSet)
	return nil
}

func InjectDisableAutoscale(p *config.KfParams) *cobra.Command {
	wire.Build(autoscaling.NewDisableAutoscaling, AppsSet)
	return nil
}

func InjectStart(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewStartCommand, AppsSet)
	return nil
}

func InjectStop(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewStopCommand, AppsSet)
	return nil
}

func InjectRestart(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewRestartCommand, AppsSet)
	return nil
}

func InjectRestage(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewRestageCommand, AppsSet)
	return nil
}

func InjectProxy(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewProxyCommand,
		AppsSet,
	)
	return nil
}

func InjectLogs(p *config.KfParams) *cobra.Command {
	wire.Build(
		clogs.NewLogsCommand,
		kflogs.NewTailer,
		config.GetKubernetes,
	)
	return nil
}

func InjectSSH(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewSSHCommand)

	return nil
}

/////////////////////////////////////
// Environment Variables Commands //
///////////////////////////////////

func InjectEnv(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewEnvCommand, AppsSet)

	return nil
}

func InjectSetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewSetEnvCommand, AppsSet)

	return nil
}

func InjectUnsetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewUnsetEnvCommand, AppsSet)

	return nil
}

////////////////
// Services //
/////////////

func provideServiceInstancesGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.ServiceInstancesGetter {
	return ki
}

func provideSecretsGetter(ki k8sclient.Interface) corev1type.SecretsGetter {
	return ki.CoreV1()
}

var ServicesSet = wire.NewSet(
	provideSecretsGetter,
	config.GetKubernetes,
	config.GetKfClient,
	provideServiceInstancesGetter,
	marketplace.NewClient,
	secrets.NewClient,
	serviceinstances.NewClient,
)

func InjectCreateService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewCreateServiceCommand,
		ServicesSet,
	)
	return nil
}

func InjectCreateUserProvidedService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewCreateUserProvidedServiceCommand,
		ServicesSet,
	)
	return nil
}

func InjectUpdateUserProvidedService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewUpdateUserProvidedServiceCommand,
		ServicesSet,
	)
	return nil
}

func InjectDeleteService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewDeleteServiceCommand,
		ServicesSet,
	)
	return nil
}

func InjectGetService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewGetServiceCommand,
	)
	return nil
}

func InjectListServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewListServicesCommand,
	)
	return nil
}

func InjectMarketplace(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewMarketplaceCommand,
		ServicesSet,
	)
	return nil
}

// /////////////////////
// Service Bindings //
// ///////////////////
func provideServiceInstanceBindingsGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.ServiceInstanceBindingsGetter {
	return ki
}

var ServiceBindingsSet = wire.NewSet(
	AppsSet,
	provideSecretsGetter,
	provideServiceInstanceBindingsGetter,
	secrets.NewClient,
	serviceinstancebindings.NewClient,
)

func InjectBindService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewBindServiceCommand,
		ServiceBindingsSet,
	)
	return nil
}

func InjectBindRouteService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewBindRouteServiceCommand,
		ServiceBindingsSet,
	)
	return nil
}

func InjectListBindings(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewListBindingsCommand,
	)
	return nil
}

func InjectUnbindService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewUnbindServiceCommand,
		ServiceBindingsSet,
	)
	return nil
}

func InjectUnbindRouteService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewUnbindRouteServiceCommand,
		ServiceBindingsSet,
	)
	return nil
}

func InjectVcapServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewVcapServicesCommand,
		config.GetKubernetes,
	)
	return nil
}

func InjectFixOrphanedBindingsCommand(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewFixOrphanedBindingsCommand,
		ServiceBindingsSet,
	)
	return nil
}

///////////////////////
// Service Brokers  //
/////////////////////

var serviceBrokerSet = wire.NewSet(
	provideSecretsGetter,
	config.GetKubernetes,
	clusterbrokerclient.NewClient,
	namespacedbrokerclient.NewClient,
	config.GetKfClient,
	secrets.NewClient,
)

func InjectCreateServiceBroker(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebrokerscmd.NewCreateServiceBrokerCommand,
		serviceBrokerSet,
	)
	return nil
}

func InjectDeleteServiceBroker(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebrokerscmd.NewDeleteServiceBrokerCommand,
		serviceBrokerSet,
	)
	return nil
}

// ///////////////
// Buildpacks //
// /////////////
func provideRemoteImageFetcher() buildpacks.RemoteImageFetcher {
	return remote.Image
}

func InjectBuildpacksClient(p *config.KfParams) buildpacks.Client {
	wire.Build(
		buildpacks.NewClient,
		provideRemoteImageFetcher,
	)
	return nil
}

func InjectWrapV2Buildpack(p *config.KfParams) *cobra.Command {
	wire.Build(cbuildpacks.NewWrapV2BuildpackCommand)
	return nil
}

func InjectBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		cbuildpacks.NewBuildpacksCommand,
		InjectBuildpacksClient,
	)
	return nil
}

func InjectStacks(p *config.KfParams) *cobra.Command {
	wire.Build(cbuildpacks.NewStacksCommand)
	return nil
}

////////////////////
// Spaces Command //
////////////////////

var SpacesSet = wire.NewSet(config.GetKfClient, config.GetKubernetes, provideKfSpaces, spaces.NewClient)

func provideKfSpaces(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.SpacesGetter {
	return ki
}

func InjectSpaces(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewListSpacesCommand)

	return nil
}

func InjectSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewGetSpaceCommand)

	return nil
}

func InjectCreateSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewCreateSpaceCommand, SpacesSet)

	return nil
}

func InjectDeleteSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewDeleteSpaceCommand)

	return nil
}

func InjectConfigSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewConfigSpaceCommand, SpacesSet)

	return nil
}

func InjectSetSpaceRole(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewSetSpaceRoleCommand, SpacesSet)

	return nil
}

func InjectSpaceUsers(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewSpaceUsersCommand, SpacesSet)

	return nil
}

func InjectUnsetSpaceRole(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewUnsetSpaceRoleCommand, SpacesSet)

	return nil
}

func InjectDomains(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewDomainsCommand, SpacesSet)

	return nil
}

func InjectTarget(p *config.KfParams) *cobra.Command {
	wire.Build(NewTargetCommand, SpacesSet)

	return nil
}

//////////////////////////
// ConfigMaps / Cluster //
//////////////////////////

var ConfigMapsSet = wire.NewSet(config.GetKubernetes, provideConfigMapsGetter, configmaps.NewClient)

func provideConfigMapsGetter(ki k8sclient.Interface) v1.ConfigMapsGetter {
	return ki.CoreV1()
}

func InjectConfigCluster(p *config.KfParams) *cobra.Command {
	wire.Build(ccluster.NewConfigClusterCommand, ConfigMapsSet)

	return nil
}

////////////
// Routes //
///////////

func InjectRoutes(p *config.KfParams) *cobra.Command {
	wire.Build(croutes.NewRoutesCommand)
	return nil
}

func InjectCreateRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewCreateRouteCommand,
		routes.NewClient,
		config.GetKfClient,
	)
	return nil
}

func InjectDeleteRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewDeleteRouteCommand,
		routes.NewClient,
		AppsSet,
	)
	return nil
}

func InjectDeleteOrphanedRoutes(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewDeleteOrphanedRoutesCommand,
		routes.NewClient,
		config.GetKfClient,
	)
	return nil
}

func InjectMapRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewMapRouteCommand,
		AppsSet,
	)
	return nil
}

func InjectUnmapRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewUnmapRouteCommand,
		AppsSet,
	)
	return nil
}

func InjectProxyRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewProxyRouteCommand,
	)
	return nil
}

////////////////////
// Builds Command //
////////////////////

var BuildsSet = wire.NewSet(
	config.GetKfClient,
	builds.TektonLoggingShim,
	provideKfBuilds,
	builds.NewClient,
	config.GetKubernetes,
)

func provideKfBuilds(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.BuildsGetter {
	return ki
}

func InjectBuilds(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewBuildsCommand)

	return nil
}

func InjectBuild(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewGetBuildCommand)

	return nil
}

func InjectBuildLogs(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewBuildLogsCommand, BuildsSet)

	return nil
}

////////////
// Tasks //
///////////

func provideTasksGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.TasksGetter {
	return ki
}

var TasksSet = wire.NewSet(
	AppsSet,
	provideTasksGetter,
	tasks.NewClient,
)

func InjectRunTask(p *config.KfParams) *cobra.Command {
	wire.Build(ctasks.NewRunTaskCommand, TasksSet)

	return nil
}

func InjectTerminateTask(p *config.KfParams) *cobra.Command {
	wire.Build(ctasks.NewTerminateTaskCommand, TasksSet)

	return nil
}

func InjectTasks(p *config.KfParams) *cobra.Command {
	wire.Build(ctasks.NewTasksCommand)

	return nil
}

///////////////////
// TaskSchedules //
///////////////////

func InjectCreateJob(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewCreateJobCommand)

	return nil
}

func InjectRunJob(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewRunJobCommand)

	return nil
}

func InjectScheduleJob(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewScheduleJobCommand)

	return nil
}

func InjectListJobs(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewListJobsCommand)

	return nil
}

func InjectListJobSchedules(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewListJobSchedulesCommand)

	return nil
}

func InjectJobHistory(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewJobHistoryCommand)

	return nil
}

func InjectDeleteJob(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewDeleteJobCommand)

	return nil
}

func InjectDeleteJobSchedule(p *config.KfParams) *cobra.Command {
	wire.Build(ctaskschedules.NewDeleteJobScheduleCommand)

	return nil
}

///////////////////////
// Other commands
///////////////////////

func InjectDependencyCommand(p *config.KfParams) *cobra.Command {
	wire.Build(dependencies.NewDependencyCommand)

	return nil
}

func InjectExportToK8sCommand(p *config.KfParams) *cobra.Command {
	wire.Build(exporttok8s.NewExportToK8s)

	return nil
}

//////////////////////////
// NetworkPolicy commands
//////////////////////////

func InjectNetworkPolicies(p *config.KfParams) *cobra.Command {
	wire.Build(cnetworkpolicies.NewListCommand)

	return nil
}

func InjectDeleteNetworkPolicies(p *config.KfParams) *cobra.Command {
	wire.Build(cnetworkpolicies.NewDeleteCommand)

	return nil
}

func InjectDescribeNetworkPolicy(p *config.KfParams) *cobra.Command {
	wire.Build(cnetworkpolicies.NewDescribeCommand)

	return nil
}
