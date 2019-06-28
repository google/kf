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
	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/buildpacks"
	"github.com/google/kf/pkg/kf/builds"
	capps "github.com/google/kf/pkg/kf/commands/apps"
	cbuildpacks "github.com/google/kf/pkg/kf/commands/buildpacks"
	"github.com/google/kf/pkg/kf/commands/config"
	cquotas "github.com/google/kf/pkg/kf/commands/quotas"
	croutes "github.com/google/kf/pkg/kf/commands/routes"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	cspaces "github.com/google/kf/pkg/kf/commands/spaces"
	kflogs "github.com/google/kf/pkg/kf/logs"
	"github.com/google/kf/pkg/kf/quotas"
	"github.com/google/kf/pkg/kf/routes"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/google/kf/pkg/kf/systemenvinjector"
	"github.com/google/wire"
	"github.com/knative/build/pkg/logs"
	"github.com/poy/kontext"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func provideSrcImageBuilder() capps.SrcImageBuilder {
	return capps.SrcImageBuilderFunc(kontext.BuildImage)
}

func provideBuildTailer() builds.BuildTailer {
	return builds.BuildTailerFunc(logs.Tail)
}

///////////////////
// App Commands //
/////////////////

var AppsSet = wire.NewSet(
	apps.NewClient,
	config.GetServingClient,
	provideSystemEnvInjector,
)

func InjectPush(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewPushCommand,
		kf.NewPusher,
		kf.NewLogTailer,
		kf.NewDeployer,
		config.GetBuildClient,
		builds.NewClient,
		provideSrcImageBuilder,
		provideBuildTailer,
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

func InjectProxy(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewProxyCommand,
		AppsSet,
		kf.NewIstioClient,
		config.GetKubernetes,
	)
	return nil
}

func InjectLogs(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewLogsCommand,
		kflogs.NewTailer,
		provideCoreV1,
	)
	return nil
}

func provideCoreV1(p *config.KfParams) corev1.CoreV1Interface {
	return config.GetKubernetes(p).CoreV1()
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

func provideSystemEnvInjector(p *config.KfParams) systemenvinjector.SystemEnvInjectorInterface {
	wire.Build(
		systemenvinjector.NewSystemEnvInjector,
		servicebindings.NewClient,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

////////////////
// Services //
/////////////
func InjectCreateService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewCreateServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func InjectDeleteService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewDeleteServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func InjectGetService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewGetServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func InjectListServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewListServicesCommand,
		config.GetSvcatApp,
	)
	return nil
}

func InjectMarketplace(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewMarketplaceCommand,
		config.GetSvcatApp,
	)
	return nil
}

///////////////////////
// Service Bindings //
/////////////////////
func InjectBindingService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewBindServiceCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func InjectListBindings(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewListBindingsCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func InjectUnbindService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewUnbindServiceCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func InjectVcapServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewVcapServicesCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
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
		config.GetBuildClient,
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
	wire.Build(cspaces.NewListSpacesCommand, SpacesSet)

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

////////////////////
// Quotas Command //
////////////////////

var QuotasSet = wire.NewSet(config.GetKubernetes, provideQuotaGetter, quotas.NewClient)

func provideQuotaGetter(ki kubernetes.Interface) v1.ResourceQuotasGetter {
	return ki.CoreV1()
}

func InjectQuotas(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewListQuotasCommand, QuotasSet)

	return nil
}

func InjectCreateQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewCreateQuotaCommand, QuotasSet)

	return nil
}

func InjectUpdateQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewUpdateQuotaCommand, QuotasSet)

	return nil
}

func InjectGetQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewGetQuotaCommand, QuotasSet)

	return nil
}

func InjectDeleteQuota(p *config.KfParams) *cobra.Command {
	wire.Build(cquotas.NewDeleteQuotaCommand, QuotasSet)

	return nil
}

////////////
// Routes //
///////////

func InjectRoutes(p *config.KfParams) *cobra.Command {
	wire.Build(
		croutes.NewRoutesCommand,
		routes.NewClient,
		config.GetKfClient,
		AppsSet,
	)
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
		config.GetKfClient,
	)
	return nil
}
