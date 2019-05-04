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
	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	cbuildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	servicebindingscmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/service-bindings"
	servicescmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
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
)

func provideSrcImageBuilder() kf.SrcImageBuilder {
	return kf.SrcImageBuilderFunc(kontext.BuildImage)
}

func provideBuildTailer() kf.BuildTailer {
	return kf.BuildTailerFunc(logs.Tail)
}

///////////////////
// App Commands //
/////////////////

func injectPush(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewPushCommand,
		kf.NewLister,
		kf.NewPusher,
		kf.NewLogTailer,
		config.GetServingClient,
		config.GetBuildClient,
		provideSrcImageBuilder,
		provideBuildTailer,
	)
	return nil
}

func injectDelete(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewDeleteCommand,
		kf.NewDeleter,
		config.GetServingClient,
	)
	return nil
}

func injectApps(p *config.KfParams) *cobra.Command {
	wire.Build(apps.NewAppsCommand, kf.NewLister, config.GetServingClient)
	return nil
}

func injectProxy(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewProxyCommand,
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

func injectEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

func injectSetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewSetEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

func injectUnsetEnv(p *config.KfParams) *cobra.Command {
	wire.Build(
		apps.NewUnsetEnvCommand,
		kf.NewLister,
		kf.NewEnvironmentClient,
		config.GetServingClient,
	)
	return nil
}

////////////////
// Services //
/////////////
func injectCreateService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewCreateServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func injectDeleteService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewDeleteServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func injectGetService(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewGetServiceCommand,
		config.GetSvcatApp,
	)
	return nil
}

func injectListServices(p *config.KfParams) *cobra.Command {
	wire.Build(
		services.NewClient,
		servicescmd.NewListServicesCommand,
		config.GetSvcatApp,
	)
	return nil
}

func injectMarketplace(p *config.KfParams) *cobra.Command {
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
func injectBindingService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewBindServiceCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func injectListBindings(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewListBindingsCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func injectUnbindService(p *config.KfParams) *cobra.Command {
	wire.Build(
		servicebindings.NewClient,
		servicebindingscmd.NewUnbindServiceCommand,
		config.GetServiceCatalogClient,
		config.GetSecretClient,
	)
	return nil
}

func injectVcapServices(p *config.KfParams) *cobra.Command {
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

func injectBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		buildpacks.NewBuildpackLister,
		cbuildpacks.NewBuildpacks,
		config.GetBuildClient,
		provideRemoteImageFetcher,
	)
	return nil
}

func injectUploadBuildpacks(p *config.KfParams) *cobra.Command {
	wire.Build(
		buildpacks.NewBuildTemplateUploader,
		cbuildpacks.NewUploadBuildpacks,
		config.GetBuildClient,
		provideBuilderCreator,
	)
	return nil
}
