package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	buildlogs "github.com/knative/build/pkg/logs"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/poy/kontext"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfgFile     string
	kubeCfgFile string
)

func getConfig() (serving.ServingV1alpha1Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgFile)
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
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgFile)
	if err != nil {
		return nil, err
	}
	client, err := build.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
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

	Supported sub-commands are:

	  kf push
	  kf delete <app>
	  kf apps

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

	lister := kf.NewLister(getConfig)
	buildLog := kf.NewLogTailer(getBuildConfig, getConfig, buildlogs.Tail)
	rootCmd.AddCommand(apps.NewDeleteCommand(p, kf.NewDeleter(getConfig)))
	rootCmd.AddCommand(apps.NewPushCommand(
		p,
		kf.NewPusher(lister, getConfig, kontext.BuildImage, buildLog)),
	)
	rootCmd.AddCommand(apps.NewAppsCommand(p, lister))

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
