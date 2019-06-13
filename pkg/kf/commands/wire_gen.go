// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package commands

import (
	"context"
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/apps"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	apps2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	buildpacks2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	quotas2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/quotas"
	servicebindings2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/service-bindings"
	services2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	spaces2 "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/spaces"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/utils"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/logs"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/quotas"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/spaces"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/systemenvinjector"
	"github.com/buildpack/lifecycle/image"
	"github.com/buildpack/pack"
	config2 "github.com/buildpack/pack/config"
	"github.com/buildpack/pack/docker"
	"github.com/buildpack/pack/fs"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/wire"
	logs2 "github.com/knative/build/pkg/logs"
	"github.com/poy/kontext"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Injectors from wire_injector.go:

func InjectPush(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	deployer := kf.NewDeployer(client)
	buildV1alpha1Interface := config.GetBuildClient(p)
	buildTailer := provideBuildTailer()
	logs := kf.NewLogTailer(buildV1alpha1Interface, servingV1alpha1Interface, buildTailer)
	pusher := kf.NewPusher(deployer, logs)
	srcImageBuilder := provideSrcImageBuilder()
	command := apps2.NewPushCommand(p, pusher, srcImageBuilder)
	return command
}

func InjectDelete(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	command := apps2.NewDeleteCommand(p, client)
	return command
}

func InjectApps(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	command := apps2.NewAppsCommand(p, client)
	return command
}

func InjectProxy(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	kubernetesInterface := config.GetKubernetes(p)
	ingressLister := kf.NewIstioClient(kubernetesInterface)
	command := apps2.NewProxyCommand(p, client, ingressLister)
	return command
}

func InjectLogs(p *config.KfParams) *cobra.Command {
	coreV1Interface := provideCoreV1(p)
	tailer := logs.NewTailer(coreV1Interface)
	command := apps2.NewLogsCommand(p, tailer)
	return command
}

func InjectEnv(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	command := apps2.NewEnvCommand(p, client)
	return command
}

func InjectSetEnv(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	command := apps2.NewSetEnvCommand(p, client)
	return command
}

func InjectUnsetEnv(p *config.KfParams) *cobra.Command {
	servingV1alpha1Interface := config.GetServingClient(p)
	systemEnvInjectorInterface := provideSystemEnvInjector(p)
	client := apps.NewClient(servingV1alpha1Interface, systemEnvInjectorInterface)
	command := apps2.NewUnsetEnvCommand(p, client)
	return command
}

func provideSystemEnvInjector(p *config.KfParams) systemenvinjector.SystemEnvInjectorInterface {
	servicecatalogV1beta1Interface := config.GetServiceCatalogClient(p)
	clientInterface := config.GetSecretClient(p)
	servicebindingsClientInterface := servicebindings.NewClient(servicecatalogV1beta1Interface, clientInterface)
	systemEnvInjectorInterface := systemenvinjector.NewSystemEnvInjector(servicebindingsClientInterface)
	return systemEnvInjectorInterface
}

func InjectCreateService(p *config.KfParams) *cobra.Command {
	sClientFactory := config.GetSvcatApp(p)
	clientInterface := services.NewClient(sClientFactory)
	command := services2.NewCreateServiceCommand(p, clientInterface)
	return command
}

func InjectDeleteService(p *config.KfParams) *cobra.Command {
	sClientFactory := config.GetSvcatApp(p)
	clientInterface := services.NewClient(sClientFactory)
	command := services2.NewDeleteServiceCommand(p, clientInterface)
	return command
}

func InjectGetService(p *config.KfParams) *cobra.Command {
	sClientFactory := config.GetSvcatApp(p)
	clientInterface := services.NewClient(sClientFactory)
	command := services2.NewGetServiceCommand(p, clientInterface)
	return command
}

func InjectListServices(p *config.KfParams) *cobra.Command {
	sClientFactory := config.GetSvcatApp(p)
	clientInterface := services.NewClient(sClientFactory)
	command := services2.NewListServicesCommand(p, clientInterface)
	return command
}

func InjectMarketplace(p *config.KfParams) *cobra.Command {
	sClientFactory := config.GetSvcatApp(p)
	clientInterface := services.NewClient(sClientFactory)
	command := services2.NewMarketplaceCommand(p, clientInterface)
	return command
}

func InjectBindingService(p *config.KfParams) *cobra.Command {
	servicecatalogV1beta1Interface := config.GetServiceCatalogClient(p)
	clientInterface := config.GetSecretClient(p)
	servicebindingsClientInterface := servicebindings.NewClient(servicecatalogV1beta1Interface, clientInterface)
	command := servicebindings2.NewBindServiceCommand(p, servicebindingsClientInterface)
	return command
}

func InjectListBindings(p *config.KfParams) *cobra.Command {
	servicecatalogV1beta1Interface := config.GetServiceCatalogClient(p)
	clientInterface := config.GetSecretClient(p)
	servicebindingsClientInterface := servicebindings.NewClient(servicecatalogV1beta1Interface, clientInterface)
	command := servicebindings2.NewListBindingsCommand(p, servicebindingsClientInterface)
	return command
}

func InjectUnbindService(p *config.KfParams) *cobra.Command {
	servicecatalogV1beta1Interface := config.GetServiceCatalogClient(p)
	clientInterface := config.GetSecretClient(p)
	servicebindingsClientInterface := servicebindings.NewClient(servicecatalogV1beta1Interface, clientInterface)
	command := servicebindings2.NewUnbindServiceCommand(p, servicebindingsClientInterface)
	return command
}

func InjectVcapServices(p *config.KfParams) *cobra.Command {
	servicecatalogV1beta1Interface := config.GetServiceCatalogClient(p)
	clientInterface := config.GetSecretClient(p)
	servicebindingsClientInterface := servicebindings.NewClient(servicecatalogV1beta1Interface, clientInterface)
	command := servicebindings2.NewVcapServicesCommand(p, servicebindingsClientInterface)
	return command
}

func InjectBuildpacksClient(p *config.KfParams) buildpacks.Client {
	buildV1alpha1Interface := config.GetBuildClient(p)
	remoteImageFetcher := provideRemoteImageFetcher()
	builderFactoryCreate := provideBuilderCreate()
	client := buildpacks.NewClient(buildV1alpha1Interface, remoteImageFetcher, builderFactoryCreate)
	return client
}

func InjectBuildpacks(p *config.KfParams) *cobra.Command {
	client := InjectBuildpacksClient(p)
	command := buildpacks2.NewBuildpacks(p, client)
	return command
}

func InjectUploadBuildpacks(p *config.KfParams) *cobra.Command {
	client := InjectBuildpacksClient(p)
	command := buildpacks2.NewUploadBuildpacks(p, client)
	return command
}

func InjectOverrider(p *config.KfParams) utils.CommandOverrideFetcher {
	kfV1alpha1Interface := config.GetKfClient(p)
	buildV1alpha1Interface := config.GetBuildClient(p)
	buildTailer := provideBuildTailer()
	srcImageBuilder := provideSrcImageBuilder()
	commandOverrideFetcher := utils.NewCommandOverrideFetcher(kfV1alpha1Interface, buildV1alpha1Interface, buildTailer, srcImageBuilder, p)
	return commandOverrideFetcher
}

func InjectSpaces(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	namespacesGetter := provideNamespaceGetter(kubernetesInterface)
	client := spaces.NewClient(namespacesGetter)
	command := spaces2.NewListSpacesCommand(p, client)
	return command
}

func InjectCreateSpace(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	namespacesGetter := provideNamespaceGetter(kubernetesInterface)
	client := spaces.NewClient(namespacesGetter)
	command := spaces2.NewCreateSpaceCommand(p, client)
	return command
}

func InjectDeleteSpace(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	namespacesGetter := provideNamespaceGetter(kubernetesInterface)
	client := spaces.NewClient(namespacesGetter)
	command := spaces2.NewDeleteSpaceCommand(p, client)
	return command
}

func InjectQuotas(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	resourceQuotasGetter := provideQuotaGetter(kubernetesInterface)
	client := quotas.NewClient(resourceQuotasGetter)
	command := quotas2.NewListQuotasCommand(p, client)
	return command
}

func InjectCreateQuota(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	resourceQuotasGetter := provideQuotaGetter(kubernetesInterface)
	client := quotas.NewClient(resourceQuotasGetter)
	command := quotas2.NewCreateQuotaCommand(p, client)
	return command
}

func InjectUpdateQuota(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	resourceQuotasGetter := provideQuotaGetter(kubernetesInterface)
	client := quotas.NewClient(resourceQuotasGetter)
	command := quotas2.NewUpdateQuotaCommand(p, client)
	return command
}

func InjectGetQuota(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	resourceQuotasGetter := provideQuotaGetter(kubernetesInterface)
	client := quotas.NewClient(resourceQuotasGetter)
	command := quotas2.NewGetQuotaCommand(p, client)
	return command
}

func InjectDeleteQuota(p *config.KfParams) *cobra.Command {
	kubernetesInterface := config.GetKubernetes(p)
	resourceQuotasGetter := provideQuotaGetter(kubernetesInterface)
	client := quotas.NewClient(resourceQuotasGetter)
	command := quotas2.NewDeleteQuotaCommand(p, client)
	return command
}

// wire_injector.go:

func provideSrcImageBuilder() apps2.SrcImageBuilder {
	return apps2.SrcImageBuilderFunc(kontext.BuildImage)
}

func provideBuildTailer() kf.BuildTailer {
	return kf.BuildTailerFunc(logs2.Tail)
}

var AppsSet = wire.NewSet(apps.NewClient, config.GetServingClient, provideSystemEnvInjector)

func provideCoreV1(p *config.KfParams) v1.CoreV1Interface {
	return config.GetKubernetes(p).CoreV1()
}

/////////////////
// Buildpacks //
///////////////
func provideRemoteImageFetcher() buildpacks.RemoteImageFetcher {
	return remote.Image
}

func provideBuilderCreate() buildpacks.BuilderFactoryCreate {
	return func(flags pack.CreateBuilderFlags) error {
		factory, err := image.NewFactory()
		if err != nil {
			return err
		}

		dockerClient, err := docker.New()
		if err != nil {
			return err
		}

		cfg, err := config2.NewDefault()
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
		builderConfig, err := builderFactory.BuilderConfigFromFlags(context.Background(), flags)
		if err != nil {
			return err
		}

		if err := builderFactory.Create(builderConfig); err != nil {
			return err
		}

		return nil
	}
}

var SpacesSet = wire.NewSet(config.GetKubernetes, provideNamespaceGetter, spaces.NewClient)

func provideNamespaceGetter(ki kubernetes.Interface) v1.NamespacesGetter {
	return ki.CoreV1()
}

var QuotasSet = wire.NewSet(config.GetKubernetes, provideQuotaGetter, quotas.NewClient)

func provideQuotaGetter(ki kubernetes.Interface) v1.ResourceQuotasGetter {
	return ki.CoreV1()
}
