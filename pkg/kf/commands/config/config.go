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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kf "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	"github.com/google/kf/pkg/kf/secrets"
	"github.com/google/kf/pkg/kf/services"
	"github.com/imdario/mergo"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	svcatclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/poy/service-catalog/pkg/svcat"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	"gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	// LogHTTP enables HTTP tracing for all Kubernetes calls.
	LogHTTP bool `yaml:"logHTTP"`

	// TargetSpace caches the space specified by Namespace to prevent it from
	// being computed multiple times.
	// Prefer using GetSpaceOrDefault instead of accessing this value directly.
	TargetSpace *v1alpha1.Space `yaml:"-"`
}

// GetTargetSpaceOrDefault gets the space specified by Namespace or a default
// space with minimal configuration.
//
// This function caches a space once retrieved in CurrentSpace.
func (p *KfParams) GetTargetSpaceOrDefault() (*v1alpha1.Space, error) {
	if p.TargetSpace != nil {
		return p.TargetSpace, nil
	}

	res, err := GetKfClient(p).Spaces().Get(p.Namespace, metav1.GetOptions{})
	return p.cacheSpace(res, err)
}

// cacheSpace updates the cached space retartieved by GetSpaceOrDefault()
func (p *KfParams) cacheSpace(space *v1alpha1.Space, err error) (*v1alpha1.Space, error) {
	if err == nil {
		p.TargetSpace = space
		return space, nil
	}

	if apierrors.IsNotFound(err) {
		p.SetTargetSpaceToDefault()
		return p.TargetSpace, nil
	}

	return nil, fmt.Errorf("couldn't get the Space %q: %v", p.Namespace, err)
}

// SetTargetSpaceToDefault sets TargetSpace to the default, overwriting
// any existing values.
func (p *KfParams) SetTargetSpaceToDefault() {
	out := &v1alpha1.Space{}
	out.SetDefaults(context.Background())
	p.TargetSpace = out
}

// paramsPath gets the path we should read the config from/write it to.
func paramsPath(userProvidedPath string) string {
	if userProvidedPath != "" {
		return userProvidedPath
	}

	return path.Join(homedir.HomeDir(), ".kf")
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
	defaultParams := initKubeConfig()

	return &defaultParams
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
func GetServiceCatalogClient(p *KfParams) servicecatalogclient.Interface {
	config := getRestConfig(p)

	cs, err := servicecatalogclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to build clientset: %s", err)
	}

	return cs
}

// GetDynamicClient gets a dynamic Kubernetes client. Dynamic clients can work
// on any type of Kubernetes object, but only support the common fields.
//
// Dynamic clients can be used to get alternative representations of objects
// like tables, or traverse multiple types of object in a single pass e.g. to
// construct a tree based on OwnerReferences.
func GetDynamicClient(p *KfParams) dynamic.Interface {
	config := getRestConfig(p)

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create a dynamic client: %s", err)
	}

	return dyn
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

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if p.KubeCfgFile != "" {
		fileList := filepath.SplitList(p.KubeCfgFile)
		if len(fileList) > 1 {
			loadingRules.Precedence = fileList
		} else {
			loadingRules.ExplicitPath = p.KubeCfgFile
		}
	}

	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	restCfg, err := clientCfg.ClientConfig()
	if err != nil {
		return &rest.Config{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, fmt.Errorf("failed to build clientcmd: %s", err)
			},
		}
	}

	if p.LogHTTP {
		restCfg.WrapTransport = NewLoggingRoundTripper
	}

	return restCfg
}

func initKubeConfig() KfParams {
	p := KfParams{}

	p.KubeCfgFile = clientcmd.RecommendedHomeFile

	// Override path to Kube config if env var is set
	if configPathEnv, ok := os.LookupEnv(clientcmd.RecommendedConfigPathEnvVar); ok {
		p.KubeCfgFile = configPathEnv
	}

	return p
}
