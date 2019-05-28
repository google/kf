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
	"context"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/apps"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	capps "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	cbuildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	servicebindingscmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/service-bindings"
	servicescmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	cspaces "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/spaces"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/utils"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/spaces"
	"github.com/buildpack/lifecycle/image"
	"github.com/buildpack/pack"
	packconfig "github.com/buildpack/pack/config"
	"github.com/buildpack/pack/docker"
	"github.com/buildpack/pack/fs"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/wire"
	"github.com/knative/build/pkg/logs"
	"github.com/poy/kontext"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func provideSrcImageBuilder() capps.SrcImageBuilder {
	return capps.SrcImageBuilderFunc(kontext.BuildImage)
}

func provideBuildTailer() kf.BuildTailer {
	return kf.BuildTailerFunc(logs.Tail)
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
		kf.NewLister,
		kf.NewPusher,
		kf.NewLogTailer,
		kf.NewDeployer,
		config.GetServingClient,
		config.GetBuildClient,
		provideSrcImageBuilder,
		provideBuildTailer,
		provideSystemEnvInjector,
	)
	return nil
}

func InjectDelete(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewDeleteCommand,
		AppsSet,
	)
	return nil
}

func InjectApps(p *config.KfParams) *cobra.Command {
	wire.Build(capps.NewAppsCommand, kf.NewLister, config.GetServingClient)
	return nil
}

func InjectProxy(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewProxyCommand,
		kf.NewLister,
		kf.NewIstioClient,
		config.GetServingClient,
		config.GetKubernetes,
	)
	return nil
}

/////////////////////////////////////
// Environment Variables Commands //
///////////////////////////////////

func InjectEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

func InjectSetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewSetEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

func InjectUnsetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		capps.NewUnsetEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

func provideSystemEnvInjector(p *config.KfParams) kf.SystemEnvInjectorInterface {
	wire.Build(
		kf.NewSystemEnvInjector,
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

func provideBuilderCreator() buildpacks.BuilderCreator {
	return buildpacks.NewBuilderCreator(func(flags pack.CreateBuilderFlags) error {
		factory, err := image.NewFactory()
		if err != nil {
			return err
		}

		dockerClient, err := docker.New()
		if err != nil {
			return err
		}

		cfg, err := packconfig.NewDefault()
		if err != nil {
			return err
		}
		builderFactory := pack.BuilderFactory{
			FS:     &fs.FS{},
			Config: cfg,
			Fetcher: &pack.ImageFetcher{
				Factory: factory,
				Docker:  dockerClient,
			},
		}
		builderConfig, err := builderFactory.BuilderConfigFromFlags(
			context.Background(),
			flags,
		)
		if err != nil {
			return err
		}

		if err := builderFactory.Create(builderConfig); err != nil {
			return err
		}

		return nil
	})
}

func InjectBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		buildpacks.NewBuildpackLister,
		cbuildpacks.NewBuildpacks,
		config.GetBuildClient,
		provideRemoteImageFetcher,
	)
	return nil
}

func InjectUploadBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		buildpacks.NewBuildTemplateUploader,
		cbuildpacks.NewUploadBuildpacks,
		config.GetBuildClient,
		provideBuilderCreator,
	)
	return nil
}

////////////////////////
// Command Overrider //
//////////////////////
func InjectOverrider(p *config.KfParams) utils.CommandOverrideFetcher {
	wire.Build(
		utils.NewCommandOverrideFetcher,
		config.GetBuildClient,
		config.GetKfClient,
		provideBuildTailer,
		provideSrcImageBuilder,
	)
	return nil
}

////////////////////
// Spaces Command //
////////////////////

var SpacesSet = wire.NewSet(config.GetKubernetes, provideNamespaceGetter, spaces.NewClient)

func provideNamespaceGetter(ki kubernetes.Interface) v1.NamespacesGetter {
	return ki.CoreV1()
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
