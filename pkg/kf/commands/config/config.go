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

	kf "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/secrets"
	"github.com/google/kf/pkg/kf/services"
	"github.com/imdario/mergo"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
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

// KfParams stores everything needed to interact with the user and Knative.
type KfParams struct {
	// Config holds the path to the configuration.
	// This field isn't serialized when the config is saved.
	Config string `yaml:"-"`

	// Namespace holds the namespace kf should connect to by default.
	Namespace string `yaml:"space"`

	// KubeCfgFile holds the path to the kubeconfig.
	KubeCfgFile string `yaml:"kubeconfig"`
}

// paramsPath gets the path we should read the config from/write it to.
func paramsPath(userProvidedPath string) string {
	if userProvidedPath != "" {
		return userProvidedPath
	}

	// Kf shouldn't fail if we can't find the user's home directory, instead
	// use the current working directory.
	base := "."
	if home, err := homedir.Dir(); err == nil {
		base = home
	}

	return path.Join(base, ".kf")
}

// NewKfParamsFromFile reads the config from the specified config path or the
// default path. If the path is the default and the file doesn't yet exist, then
// this function does nothing.
func NewKfParamsFromFile(cfgPath string) (*KfParams, error) {
	configWasOverridden := cfgPath != ""
	cfgPath = paramsPath(cfgPath)

	contents, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		switch {
		case configWasOverridden:
			return nil, err
		case os.IsNotExist(err):
			return &KfParams{}, nil
		default:
			return nil, err
		}
	}

	newParams := &KfParams{}
	if err := yaml.Unmarshal(contents, newParams); err != nil {
		return nil, err
	}

	return newParams, nil
}

// NewDefaultKfParams creates a KfParams with default values.
func NewDefaultKfParams() *KfParams {
	defaultParams := &KfParams{
		Namespace: "default",
	}

	initKubeConfig(defaultParams)

	return defaultParams
}

// Write writes the current configuration to the path specified by the
// user or the default path.
func Write(cfgPath string, config *KfParams) error {
	configPath := paramsPath(cfgPath)

	contents, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, contents, 0664)
}

// Load reads the config at the given path (or the default config path if not
// provided), and merges the values with the defaults and overrides.
func Load(cfgPath string, overrides *KfParams) (*KfParams, error) {
	params, err := NewKfParamsFromFile(cfgPath)
	if err != nil {
		return nil, err
	}

	out := &KfParams{}
	mergo.Merge(out, overrides)
	mergo.Merge(out, params)
	mergo.Merge(out, NewDefaultKfParams())

	return out, nil
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
