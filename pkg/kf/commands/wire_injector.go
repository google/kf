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

//+build wireinject

package commands

import (
	"github.com/google/go-containerregistry/pkg/v1/remote"
	kfv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	scv1beta1 "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/typed/servicecatalog/v1beta1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/buildpacks"
	capps "github.com/google/kf/pkg/kf/commands/apps"
	cbuildpacks "github.com/google/kf/pkg/kf/commands/buildpacks"
	cbuilds "github.com/google/kf/pkg/kf/commands/builds"
	ccompletion "github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	cquotas "github.com/google/kf/pkg/kf/commands/quotas"
	croutes "github.com/google/kf/pkg/kf/commands/routes"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	servicebrokerscmd "github.com/google/kf/pkg/kf/commands/service-brokers"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	cspaces "github.com/google/kf/pkg/kf/commands/spaces"
	"github.com/google/kf/pkg/kf/istio"
	kflogs "github.com/google/kf/pkg/kf/logs"
	"github.com/google/kf/pkg/kf/marketplace"
	"github.com/google/kf/pkg/kf/routeclaims"
	"github.com/google/kf/pkg/kf/routes"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	clusterbrokerclient "github.com/google/kf/pkg/kf/service-brokers/cluster"
	namespacedbrokerclient "github.com/google/kf/pkg/kf/service-brokers/namespaced"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/sources"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/google/wire"
	"github.com/poy/kontext"
	"github.com/spf13/cobra"
)

func provideSrcImageBuilder() capps.SrcImageBuilder {
	return capps.SrcImageBuilderFunc(kontext.BuildImageWithFilter)
}

///////////////////
// App Commands //
/////////////////

var AppsSet = wire.NewSet(
	SourcesSet,
	provideAppsGetter,
	apps.NewClient,
	apps.NewPusher,
)

func provideAppsGetter(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.AppsGetter {
	return ki
}

func InjectPush(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewPushCommand,
		provideSrcImageBuilder,
		servicebindings.NewClient,
		config.GetServiceCatalogClient,
		AppsSet,
	)
	return nil
}

func InjectDelete(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewDeleteCommand, AppsSet)

	return nil
}

func InjectApps(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewAppsCommand, AppsSet)

	return nil
}

func InjectGetApp(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewGetAppCommand, AppsSet)

	return nil
}

func InjectScale(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewScaleCommand, AppsSet)
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
		istio.NewIstioClient,
	)
	return nil
}

func InjectLogs(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewLogsCommand,
		kflogs.NewTailer,
		config.GetKubernetes,
	)
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

func provideServiceInstancesGetter(sc servicecatalogclient.Interface) scv1beta1.ServiceInstancesGetter {
	return sc.ServicecatalogV1beta1()
}

var ServicesSet = wire.NewSet(
	provideServiceInstancesGetter,
	config.GetServiceCatalogClient,
	config.GetSvcatApp,
	marketplace.NewClient,
	services.NewClient,
)

func InjectCreateService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewCreateServiceCommand,
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
		ServicesSet,
	)
	return nil
}

func InjectListServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicescmd.NewListServicesCommand,
		ServicesSet,
		AppsSet,
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

///////////////////////
// Service Bindings //
/////////////////////
func InjectBindingService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewBindServiceCommand,
		AppsSet,
	)
	return nil
}

func InjectListBindings(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewListBindingsCommand,
		config.GetServiceCatalogClient,
	)
	return nil
}

func InjectUnbindService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindingscmd.NewUnbindServiceCommand,
		AppsSet,
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

///////////////////////
// Service Brokers  //
/////////////////////

func provideClusterServiceBrokerGetter(sc servicecatalogclient.Interface) scv1beta1.ClusterServiceBrokersGetter {
	return sc.ServicecatalogV1beta1()
}

func provideServiceBrokerGetter(sc servicecatalogclient.Interface) scv1beta1.ServiceBrokersGetter {
	return sc.ServicecatalogV1beta1()
}

var serviceBrokerSet = wire.NewSet(
	config.GetServiceCatalogClient,
	provideClusterServiceBrokerGetter,
	provideServiceBrokerGetter,
	clusterbrokerclient.NewClient,
	namespacedbrokerclient.NewClient,
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

/////////////////
// Buildpacks //
///////////////
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

func InjectBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		cbuildpacks.NewBuildpacksCommand,
		InjectBuildpacksClient,
	)
	return nil
}

func InjectStacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		cbuildpacks.NewStacksCommand,
		InjectBuildpacksClient,
	)
	return nil
}

////////////////////
// Spaces Command //
////////////////////

var SpacesSet = wire.NewSet(config.GetKfClient, provideKfSpaces, spaces.NewClient)

func provideKfSpaces(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.SpacesGetter {
	return ki
}

func InjectSpaces(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewListSpacesCommand, config.GetDynamicClient, config.GetTableClient)

	return nil
}

func InjectSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewGetSpaceCommand, SpacesSet)

	return nil
}

func InjectCreateSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewCreateSpaceCommand, SpacesSet)

	return nil
}

func InjectDeleteSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewDeleteSpaceCommand, SpacesSet)

	return nil
}

func InjectConfigSpace(p *config.KfParams) *cobra.Command {
	wire.Build(cspaces.NewConfigSpaceCommand, SpacesSet)

	return nil
}

////////////////////
// Quotas Command //
////////////////////

func InjectUpdateQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewUpdateQuotaCommand, SpacesSet)

	return nil
}

func InjectGetQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewGetQuotaCommand, SpacesSet)

	return nil
}

func InjectDeleteQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewDeleteQuotaCommand, SpacesSet)

	return nil
}

////////////
// Routes //
///////////

func InjectRoutes(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewRoutesCommand,
		routes.NewClient,
		routeclaims.NewClient,
		AppsSet,
	)
	return nil
}

func InjectCreateRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewCreateRouteCommand,
		routeclaims.NewClient,
		config.GetKfClient,
	)
	return nil
}

func InjectDeleteRoute(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewDeleteRouteCommand,
		routeclaims.NewClient,
		AppsSet,
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
		istio.NewIstioClient,
		config.GetKubernetes,
	)
	return nil
}

////////////////////
// Builds Command //
////////////////////

var SourcesSet = wire.NewSet(
	config.GetKfClient,
	sources.TektonLoggingShim,
	provideKfSources,
	sources.NewClient,
	config.GetTektonClient,
	config.GetKubernetes,
)

func provideKfSources(ki kfv1alpha1.KfV1alpha1Interface) kfv1alpha1.SourcesGetter {
	return ki
}

func InjectBuilds(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewBuildsCommand, config.GetDynamicClient, config.GetTableClient)

	return nil
}

func InjectBuild(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewGetBuildCommand, config.GetDynamicClient)

	return nil
}

func InjectBuildLogs(p *config.KfParams) *cobra.Command {
	wire.Build(cbuilds.NewBuildLogsCommand, SourcesSet)

	return nil
}

///////////////////////
// Completion commands
///////////////////////

func InjectNamesCommand(p *config.KfParams) *cobra.Command {
	wire.Build(ccompletion.NewNamesCommand, config.GetDynamicClient)

	return nil
}
