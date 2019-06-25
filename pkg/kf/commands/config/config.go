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

package config

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	kf "github.com/GoogleCloudPlatform/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/secrets"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/imdario/mergo"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	networking "github.com/knative/pkg/client/clientset/versioned/typed/istio/v1alpha3"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"
	svcatclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"
	scv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/poy/service-catalog/pkg/svcat"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	"gopkg.in/yaml.v2"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	defaultConfigPath = ""
)

func init() {
	// kf shouldn't fail if we can't find the user's home directory, use the PWD
	// instead
	home, err := homedir.Dir()
	if err != nil {
		defaultConfigPath = path.Join(".", ".kf")
	} else {
		defaultConfigPath = path.Join(home, ".kf")
	}
}

// KfParams stores everything needed to interact with the user and Knative.
type KfParams struct {
	// Config holds the path to the configuration.
	Config string `yaml:"-"`

	// Namespace holds the namespace kf should connect to by default.
	Namespace string `yaml:"space"`

	// KubeCfgFile holds the path to the kubeconfig.
	KubeCfgFile string `yaml:"kubeconfig"`
}

// ConfigPath gets the path we should read the config from/write it to.
func (p *KfParams) ConfigPath() string {
	if p.Config != "" {
		return p.Config
	}

	return defaultConfigPath
}

// ReadConfig reads the config from the specified config path or the default
// path. If the path is the default and the file doesn't yet exist, then
// this function does nothing.
func (p *KfParams) ReadConfig() error {
	userSpecifiedConfig := p.Config != ""
	configPath := p.ConfigPath()

	contents, err := ioutil.ReadFile(configPath)
	if err != nil {
		switch {
		case userSpecifiedConfig:
			return err
		case os.IsNotExist(err):
			return nil
		default:
			return err
		}
	}

	newParams := &KfParams{}
	if err := yaml.Unmarshal(contents, newParams); err != nil {
		return err
	}

	if err := mergo.Merge(p, newParams); err != nil {
		return err
	}

	return nil
}

// WriteConfig writes the current configuration to the path specified by the
// user or the default path.
func (p *KfParams) WriteConfig() error {
	configPath := p.ConfigPath()

	contents, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, contents, 0664)
}

// GetServingClient returns a Serving interface.
func GetServingClient(p *KfParams) serving.ServingV1alpha1Interface {
	config := getRestConfig(p)
	client, err := serving.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a Serving client: %s", err)
	}
	return client
}

// GetBuildClient returns a Build interface.
func GetBuildClient(p *KfParams) build.BuildV1alpha1Interface {
	config := getRestConfig(p)
	client, err := build.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a Build client: %s", err)
	}
	return client
}

// GetKubernetes returns a K8s client.
func GetKubernetes(p *KfParams) k8sclient.Interface {
	config := getRestConfig(p)
	c, err := k8sclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a K8s client: %s", err)
	}
	return c
}

// GetNetworkingClient returns a Networking interface.
func GetNetworkingClient(p *KfParams) networking.NetworkingV1alpha3Interface {
	config := getRestConfig(p)
	client, err := networking.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a Networking client: %s", err)
	}
	return client
}

// GetKfClient returns a kf client.
func GetKfClient(p *KfParams) kf.KfV1alpha1Interface {
	config := getRestConfig(p)
	c, err := kf.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a kf client: %s", err)
	}
	return c
}

// GetServingClient returns a secrets Client.
func GetSecretClient(p *KfParams) secrets.ClientInterface {
	return secrets.NewClient(GetKubernetes(p))
}

// GetServiceCatalogClient returns a ServiceCatalogClient.
func GetServiceCatalogClient(p *KfParams) scv1beta1.ServicecatalogV1beta1Interface {
	config := getRestConfig(p)

	cs, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to build clientset: %s", err)
	}

	return cs.ServicecatalogV1beta1()
}

// GetSvcatApp returns a SvcatClient.
func GetSvcatApp(p *KfParams) services.SClientFactory {
	return func(namespace string) servicecatalog.SvcatClient {
		config := getRestConfig(p)

		k8sClient, err := k8sclient.NewForConfig(config)
		if err != nil {
			log.Fatalf("failed to create a K8s client: %s", err)
		}

		catalogClient, err := svcatclient.NewForConfig(config)
		if err != nil {
			log.Fatalf("failed to create a svcatclient: %s", err)
		}

		c, err := svcat.NewApp(k8sClient, catalogClient, namespace)
		if err != nil {
			log.Fatalf("failed to create a svcat App: %s", err)
		}
		return c
	}
}

func getRestConfig(p *KfParams) *rest.Config {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config
	}

	initKubeConfig(p)
	c, err := clientcmd.BuildConfigFromFlags("", p.KubeCfgFile)
	if err != nil {
		log.Fatalf("failed to build clientcmd: %s", err)
	}
	return c
}

func initKubeConfig(p *KfParams) {
	if p.KubeCfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("failed to load kubectl config: %s", err)
		}
		p.KubeCfgFile = filepath.Join(home, ".kube", "config")
	}
}
