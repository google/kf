package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	buildlogs "github.com/knative/build/pkg/logs"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/poy/kontext"
	"github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"
	svcatclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"
	scv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/poy/service-catalog/pkg/svcat"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
	k8sclient "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfgFile     string
	kubeCfgFile string
)

func getRestConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeCfgFile)
}

func getConfig() (serving.ServingV1alpha1Interface, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}
	client, err := serving.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getBuildConfig() (build.BuildV1alpha1Interface, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}
	client, err := build.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getSvcatApp(namespace string) (servicecatalog.SvcatClient, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := k8sclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	catalogClient, err := svcatclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return svcat.NewApp(k8sClient, catalogClient, namespace)
}

func getServiceCatalogClient() (scv1beta1.ServicecatalogV1beta1Interface, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}

	cs, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return cs.ServicecatalogV1beta1(), nil
}

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	p := &config.KfParams{
		Output: os.Stdout,
	}

	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "kf is like cf for Knative",
		Long: `kf is like cf for Knative

kf supports the following sub-commands:

Apps:
  kf push
  kf delete <app>
  kf apps

Services:
  kf marketplace
  kf create-service
  kf delete-service
  kf service <instance-name>
  kf services
	kf bindings
	kf bind-service <app> <instance-name>
	kf unbind-service <app> <instance-name>

You can get more info by adding the --help flag to any sub-command.
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	// Model new commands after:
	// https://github.com/knative/client/blob/master/pkg/kn/commands/service_list.go
	// to take an idiomatic k8s like approach.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kf.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "default", "namespace")

	// App interaction
	lister := kf.NewLister(getConfig)
	buildLog := kf.NewLogTailer(getBuildConfig, getConfig, buildlogs.Tail)
	rootCmd.AddCommand(apps.NewDeleteCommand(p, kf.NewDeleter(getConfig)))
	rootCmd.AddCommand(apps.NewPushCommand(
		p,
		kf.NewPusher(lister, getConfig, kontext.BuildImage, buildLog)),
	)
	rootCmd.AddCommand(apps.NewAppsCommand(p, lister))

	// Environment Variables
	envClient := kf.NewEnvironmentClient(lister, getConfig)
	rootCmd.AddCommand(apps.NewEnvCommand(p, envClient))
	rootCmd.AddCommand(apps.NewSetEnvCommand(p, envClient))

	// Services
	servicesClient := services.NewClient(getSvcatApp)
	rootCmd.AddCommand(servicescmd.NewCreateServiceCommand(p, servicesClient))
	rootCmd.AddCommand(servicescmd.NewDeleteServiceCommand(p, servicesClient))
	rootCmd.AddCommand(servicescmd.NewGetServiceCommand(p, servicesClient))
	rootCmd.AddCommand(servicescmd.NewListServicesCommand(p, servicesClient))
	rootCmd.AddCommand(servicescmd.NewMarketplaceCommand(p, servicesClient))

	bindingClient := servicebindings.NewClient(getServiceCatalogClient)
	rootCmd.AddCommand(servicebindingscmd.NewBindServiceCommand(p, bindingClient))
	rootCmd.AddCommand(servicebindingscmd.NewListBindingsCommand(p, bindingClient))
	rootCmd.AddCommand(servicebindingscmd.NewUnbindServiceCommand(p, bindingClient))

	// Buildpacks
	buildpackLister := buildpacks.NewBuildpackLister(getBuildConfig, remote.Image)
	rootCmd.AddCommand(cbuildpacks.NewBuildpacks(p, buildpackLister))
	rootCmd.AddCommand(cbuildpacks.NewUploadBuildpacks(
		p,
		initBuilderCreator(),
		buildpacks.NewBuildTemplateUploader(getBuildConfig),
	))

	return rootCmd
}

func InitializeConfig() {
	cobra.OnInitialize(initKubeConfig)
}

func initKubeConfig() {
	if kubeCfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		kubeCfgFile = filepath.Join(home, ".kube", "config")
	}
}

func initBuilderCreator() *buildpacks.BuilderCreator {
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
