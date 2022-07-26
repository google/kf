// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/google/k8s-stateless-subresource/pkg/internal/apiserver"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/rest"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	cminformer "knative.dev/pkg/configmap/informer"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/environment"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"
	"knative.dev/pkg/version"
)

// Config has configuration for the API server.
type Config struct {
	// SecureServing has the configuration for securing the API server.
	SecureServing *options.SecureServingOptionsWithLoopback

	// Authentication has the configuration for authenticating the API server.
	Authentication *options.DelegatingAuthenticationOptions

	// Authorization has the configuration for authorization the API server.
	Authorization *options.DelegatingAuthorizationOptions

	// Features has the configuration for profiling the API server.
	Features *options.FeatureOptions

	// RESTConfig is the rest.Config used to communicate with the Kubernetes
	// API server. It can be nil.
	RESTConfig *rest.Config

	// Predicate which is true for paths of long-running http requests
	LongRunningFunc apirequest.LongRunningRequestCheck
}

// NewConfig returns a new config with the defaults set.
func NewConfig() Config {
	var cfg Config
	cfg.setDefaults()
	return cfg
}

// AddFlags adds flags for the various fields in the config. This is an
// optional (albeit convenient) way to populate the config.
func (c *Config) AddFlags(fs *pflag.FlagSet) {
	c.SecureServing.AddFlags(fs)
	c.Authentication.AddFlags(fs)
	c.Authorization.AddFlags(fs)
	c.Features.AddFlags(fs)
}

func (c *Config) setDefaults() {
	if c.RESTConfig == nil {
		env := new(environment.ClientConfig)
		var err error
		if c.RESTConfig, err = injection.GetRESTConfig(
			env.ServerURL,
			env.Kubeconfig,
		); err != nil {
			// NOTE: Normally we wouldn't want to kill the API server pod,
			// however given this will be ran within the cluster, it shouldn't
			// fail.
			log.Fatalf("failed to get rest config: %v", err)
		}
	}
	if c.SecureServing == nil {
		c.SecureServing = options.NewSecureServingOptions().WithLoopback()
	}
	if c.Authentication == nil {
		c.Authentication = options.NewDelegatingAuthenticationOptions()
	}
	if c.Authorization == nil {
		c.Authorization = options.NewDelegatingAuthorizationOptions()
	}
	if c.Features == nil {
		c.Features = options.NewFeatureOptions()
	}
}

func (c *Config) server(
	name string,
	codecs serializer.CodecFactory,
) (*apiserver.Server, error) {
	// Generate self-signed certs to use.
	if err := c.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost",
		nil,
		[]net.IP{net.ParseIP("127.0.0.1")},
	); err != nil {
		return nil, fmt.Errorf("failed to create self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(codecs)
	serverConfig.LongRunningFunc = c.LongRunningFunc

	// Apply SecureServing config.
	if err := c.SecureServing.ApplyTo(
		&serverConfig.SecureServing,
		&serverConfig.LoopbackClientConfig,
	); err != nil {
		return nil, fmt.Errorf("failed to merge SecureServing config: %v", err)
	}

	// Apply Authentication config.
	if err := c.Authentication.ApplyTo(
		&serverConfig.Authentication,
		serverConfig.SecureServing,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to merge Authentication config: %v", err)
	}

	// Apply Authorization config.
	if err := c.Authorization.ApplyTo(&serverConfig.Authorization); err != nil {
		return nil, fmt.Errorf("failed to merge Authorization config: %v", err)
	}

	// Don't bother with the informers.
	completedConfig := serverConfig.Complete(nil)

	genericServer, err := completedConfig.New(
		name,
		genericapiserver.NewEmptyDelegate(),
	)
	if err != nil {
		return nil, err
	}

	s := &apiserver.Server{
		GenericAPIServer: genericServer,
	}

	return s, nil
}

// EndpointConstructor is used to construct an endpoint.
type EndpointConstructor interface {
	// API returns the endpoint's API.
	API() (*runtime.Scheme, metav1.APIResource)

	// RegisterWebService is used to setup a WebService for the given
	// endpoint.
	RegisterWebService(context.Context, *restful.WebService) error
}

// EndpointConstructorStruct implements EndpointConstructor.
type EndpointConstructorStruct struct {
	// Scheme is the endpoint's Scheme.
	Scheme *runtime.Scheme

	// Resource is the endpoint's Resource.
	Resource metav1.APIResource

	// RegisterWebServiceF is used to setup a WebService for the given
	// endpoint.
	RegisterWebServiceF func(context.Context, *restful.WebService) error
}

// API implements EndpointConstructor.
func (f EndpointConstructorStruct) API() (*runtime.Scheme, metav1.APIResource) {
	return f.Scheme, f.Resource
}

// RegisterWebService implements EndpointConstructor.
func (f EndpointConstructorStruct) RegisterWebService(ctx context.Context, ws *restful.WebService) error {
	return f.RegisterWebServiceF(ctx, ws)
}

// Main invokes MainWithConfig with a nil config. It will block until the
// given context is done.
func Main(
	ctx context.Context,
	name string,
	ctor EndpointConstructor,
) {
	MainWithConfig(ctx, name, Config{}, ctor)
}

// MainWithConfig runs the generic main flow for API servers. It will block
// until the given context is done.
func MainWithConfig(
	ctx context.Context,
	name string,
	cfg Config,
	ctor EndpointConstructor,
) {
	// Set config defaults.
	cfg.setDefaults()

	// Setup injection.
	ctx = enableInjection(ctx, cfg.RESTConfig)

	// Setup logging.
	logger, atomicLevel := setupLogger(ctx, name)
	defer logger.Sync()
	ctx = logging.WithLogger(ctx, logger)

	go checkK8sClientMinimumVersion(ctx, logger)

	go func() {
		// Normally, we wouldn't want to use a XXXOrDie function in the API
		// server because the Pod dying can result in the Kubernetes API
		// server becoming unhealthy. However, in this case, the failure case
		// is if an environment variable isn't set (not something that would
		// be transient).
		cmw := sharedmain.SetupConfigMapWatchOrDie(ctx, logger)

		watchLoggingConfig(ctx, cmw, logger, atomicLevel, name)

		logger.Info("Starting configuration manager...")
		for ctx.Err() == nil {
			if err := cmw.Start(ctx.Done()); err != nil {
				logger.Warnw("failed to start configuration manager. retrying...", zap.Error(err))

				// Retry...
				time.Sleep(time.Second)
				continue
			}

			break
		}
	}()

	scheme, resource := ctor.API()

	// Setup codecs.
	codecs := serializer.NewCodecFactory(scheme)

	s, err := cfg.server(name, codecs)
	if err != nil {
		logger.Fatalw("failed to setup API Server", zap.Error(err))
	}

	// Build a register function with the context as the first argument. The
	// underlying library doesn't really use/care about a context.
	register := func(ws *restful.WebService) error {
		return ctor.RegisterWebService(ctx, ws)
	}

	// Install the endpoint.
	if err := s.InstallAPI(
		scheme,
		codecs,
		resource,
		register,
	); err != nil {
		logger.Fatalw("failed to install API", zap.Error(err))
	}

	// Run the server.
	if err := s.GenericAPIServer.PrepareRun().Run(ctx.Done()); err != nil {
		logger.Fatalw("failed to run", zap.Error(err))
	}
}

// checkK8sClientMinimumVersion is based on CheckK8sClientMinimumVersionOrDie
// without the death part. We don't have the luxury of simply killing the pod
// here because this is part of the API server.
func checkK8sClientMinimumVersion(ctx context.Context, logger *zap.SugaredLogger) {
	for ctx.Err() == nil {
		kc := kubeclient.Get(ctx)
		if err := version.CheckMinimumVersion(kc.Discovery()); err != nil {
			logger.Warnw("Version check failed... retrying...", zap.Error(err))

			// Retry...
			time.Sleep(time.Second)
			continue
		}

		return
	}
}

// watchLoggingConfig is based on WatchObservabilityConfigOrDie without the
// death part. We don't have the luxury of simply killing the pod here because
// this is part of the API server.
func watchLoggingConfig(ctx context.Context, cmw *cminformer.InformedWatcher, logger *zap.SugaredLogger, atomicLevel zap.AtomicLevel, component string) {
	for ctx.Err() == nil {
		if _, err := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace()).Get(ctx, logging.ConfigMapName(),
			metav1.GetOptions{}); err == nil {
			cmw.Watch(logging.ConfigMapName(), logging.UpdateLevelFromConfigMap(logger, atomicLevel, component))
			return
		} else if !apierrors.IsNotFound(err) {
			logger.Warnw("Error reading ConfigMap (retrying...) "+logging.ConfigMapName(), zap.Error(err))

			// Retry...
			time.Sleep(time.Second)
			continue
		}
	}
}

// enableInjection is based on EnableInjectionOrDie without the death part. We
// don't have the luxury of simply killing the pod here because this is part
// of the API server.
func enableInjection(ctx context.Context, cfg *rest.Config) context.Context {
	if ctx == nil {
		ctx = signals.NewContext()
	}
	if cfg == nil {
		// Normally we woudn't want to die in the API server, however this is
		// unlikely to happen when we are running in a pod.
		cfg = injection.ParseAndGetRESTConfigOrDie()
	}

	// Respect user provided settings, but if omitted customize the default behavior.
	if cfg.QPS == 0 {
		cfg.QPS = rest.DefaultQPS
	}
	if cfg.Burst == 0 {
		cfg.Burst = rest.DefaultBurst
	}
	ctx = injection.WithConfig(ctx, cfg)

	ctx, informers := injection.Default.SetupInformers(ctx, cfg)

	go func() {
		for ctx.Err() == nil {
			logging.FromContext(ctx).Info("Starting informers...")
			if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
				logging.FromContext(ctx).Warnw("Failed to start informers", zap.Error(err))

				// Retry...
				time.Sleep(time.Second)
				continue
			}

			return
		}
	}()

	return ctx
}

// setupLogger is based on SetupLoggerOrDie without the death part. We don't
// have the luxury of simply killing the pod here because this is part of the
// API server.
func setupLogger(ctx context.Context, component string) (*zap.SugaredLogger, zap.AtomicLevel) {
	loggingConfig, err := sharedmain.GetLoggingConfig(ctx)
	if err != nil {
		// XXX: We'll use the example logger, which spits out JSON instead of
		// dying.
		logger := zap.NewExample().Sugar()
		logger.Warnf("Error reading/parsing logging configuration (using base logger): %v", err)

		return logger, zap.NewAtomicLevel()
	}
	l, level := logging.NewLoggerFromConfig(loggingConfig, component)

	// If PodName is injected into the env vars, set it on the logger.
	// This is needed for HA components to distinguish logs from different
	// pods.
	if pn := os.Getenv("POD_NAME"); pn != "" {
		l = l.With(zap.String(logkey.Pod, pn))
	}

	return l, level
}
