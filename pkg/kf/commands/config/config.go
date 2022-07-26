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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"sync"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kf "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/injection"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/imdario/mergo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/homedir"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/yaml"
)

const (
	// EmptySpaceError is the message returned if the user hasn't targed a space.
	EmptySpaceError = "no space targeted, use 'kf target --space SPACE' to target a space"

	// SkipVersionCheckAnnotation can be set on commands to prevent the version
	// check from running.
	SkipVersionCheckAnnotation = "skip-version-check"
)

// KfParams stores user settings.
type KfParams struct {
	// Config holds the path to the configuration.
	// This field isn't serialized when the config is saved.
	Config string `json:"-"`

	// Space holds the namespace kf should connect to by default.
	Space string `json:"space"`

	// KubeCfgFile holds the path to the kubeconfig.
	KubeCfgFile string `json:"-"`

	// LogHTTP enables HTTP tracing for all Kubernetes calls.
	LogHTTP bool `json:"logHTTP"`

	// TargetSpace caches the space specified by Space to prevent it from
	// being computed multiple times.
	// Prefer using GetSpaceOrDefault instead of accessing this value
	// directly.
	TargetSpace *v1alpha1.Space `json:"-"`

	// featureFlags is a map of each feature flag name to a bool
	// indicating whether the feature is enabled or not.
	featureFlags     kfconfig.FeatureFlagToggles `json:"-"`
	featureFlagsOnce sync.Once                   `json:"-"`

	// Impersonate is the config that will be used to impersonate a user in
	// REST requests.
	Impersonate transport.ImpersonationConfig `json:"-"`
}

// ConfigFlags constructs a new ConfigFlags from KfParams that can be used
// to initialize a RestConfig in the same way that Kubectl does it.
func (p *KfParams) ConfigFlags() (flags *genericclioptions.ConfigFlags) {

	// The ConfigFlags object handles overiding kubeconfig paths in a consistent
	// way for all Kubernetes clients. It will correctly choose when to parse
	// environment variables, etc.
	flags = genericclioptions.NewConfigFlags(false)

	// Override default values based on flags Kf supplies.
	flags.Namespace = &p.Space
	flags.KubeConfig = &p.KubeCfgFile

	return
}

// FeatureFlags returns a map of each feature flag name to a bool
// indicating whether hte feature is enabled or not.
func (p *KfParams) FeatureFlags(ctx context.Context) kfconfig.FeatureFlagToggles {
	p.featureFlagsOnce.Do(func() {
		logger := logging.FromContext(ctx)
		kubeClient := kubeclient.Get(ctx)
		ns, err := kubeClient.
			CoreV1().
			Namespaces().
			Get(ctx, v1alpha1.KfNamespace, metav1.GetOptions{})
		if err != nil {
			logger.Warnf(
				"Error getting %s namespace for CLI warnings: %s",
				v1alpha1.KfNamespace,
				err,
			)
			return
		}

		featureFlagsMap := make(map[string]bool)
		if featureFlagJSON, ok := ns.Annotations[v1alpha1.FeatureFlagsAnnotation]; ok {
			if err := json.Unmarshal([]byte(featureFlagJSON), &featureFlagsMap); err != nil {
				logger.Warnf("Invalid feature flag config: %v", err)
				return
			}
		} else {
			logger.Warn("Unable to read feature flags from server.")
			return
		}
		p.featureFlags = featureFlagsMap
	})

	return p.featureFlags
}

func suggestBadSpaceNextActions() {
	utils.SuggestNextAction(utils.NextAction{
		Description: "List known spaces",
		Commands: []string{
			"kf spaces",
			"kubectl get spaces",
		},
	})
	utils.SuggestNextAction(utils.NextAction{
		Description: "Target a space",
		Commands: []string{
			"kf target -s SPACENAME",
		},
	})

}

// ValidateSpaceTargeted returns an error if a Space isn't targeted.
func (p *KfParams) ValidateSpaceTargeted() error {
	if p.Space == "" {
		suggestBadSpaceNextActions()
		return errors.New(EmptySpaceError)
	}
	return nil
}

// GetTargetSpace gets the space specified by Space. If the space
// doesn't exist, an error is returned
//
// This function caches a space once retrieved in TargetSpace.
func (p *KfParams) GetTargetSpace(ctx context.Context) (*v1alpha1.Space, error) {
	if p.TargetSpace != nil {
		return p.TargetSpace, nil
	}

	res, err := GetKfClient(p).Spaces().Get(ctx, p.Space, metav1.GetOptions{})
	return p.cacheSpace(res, err)
}

// cacheSpace updates the cached space retrieved by GetTargetSpace()
func (p *KfParams) cacheSpace(space *v1alpha1.Space, err error) (*v1alpha1.Space, error) {
	switch {
	case err == nil:
		p.TargetSpace = space
		return space, nil

	case apierrors.IsNotFound(err):
		suggestBadSpaceNextActions()
		return nil, fmt.Errorf("Space %q doesn't exist", p.Space)

	default:
		return nil, fmt.Errorf("couldn't get the Space %q: %v", p.Space, err)
	}
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
	defaultParams := KfParams{}

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

var (
	kubernetesOnce   sync.Once
	kubernetesClient k8sclient.Interface
)

// GetKubernetes returns a K8s client.
func GetKubernetes(p *KfParams) k8sclient.Interface {
	kubernetesOnce.Do(func() {
		config := getRestConfig(p)
		c, err := k8sclient.NewForConfig(config)
		if err != nil {
			log.Fatalf("failed to create a K8s client: %s", err)
		}
		kubernetesClient = c
	})
	return kubernetesClient
}

var (
	kfOnce   sync.Once
	kfClient kf.KfV1alpha1Interface
)

// GetKfClient returns a kf client.
func GetKfClient(p *KfParams) kf.KfV1alpha1Interface {
	kfOnce.Do(func() {
		config := getRestConfig(p)
		c, err := kf.NewForConfig(config)
		if err != nil {
			log.Fatalf("failed to create a kf client: %s", err)
		}
		kfClient = c
	})
	return kfClient
}

func getRestConfig(p *KfParams) *rest.Config {
	restCfg, err := p.ConfigFlags().ToRESTConfig()
	if err != nil {
		return &rest.Config{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, fmt.Errorf("failed to build clientcmd: %s", err)
			},
		}
	}

	// NOTE: because Kf uses wire to initialize components on startup, any
	// RoundTrippers added here MUST be reactive to changes in KfParams rather
	// than assuming the initial configuration passed is the final one.

	restCfg.Wrap(LoggingRoundTripperWrapper(p))

	restCfg.Wrap(NewImpersonatingRoundTripperWrapper(p))

	return restCfg
}

// SetupInjection sets up the injection context.
// XXX: This function is only necessary while we have a huge dependency on
// KfParams. Once that is removed, this function will no longer be necessary.
func SetupInjection(ctx context.Context, p *KfParams) context.Context {
	return injection.WithInjection(ctx, getRestConfig(p))
}
