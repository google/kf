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

package reconciler

import (
	"context"

	kfclientset "github.com/google/kf/pkg/client/clientset/versioned"
	kfscheme "github.com/google/kf/pkg/client/clientset/versioned/scheme"
	kfclient "github.com/google/kf/pkg/client/injection/client"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	sharedclientset "knative.dev/pkg/client/clientset/versioned"
	sharedclient "knative.dev/pkg/client/injection/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/injection/clients/kubeclient"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
)

// Base implements the core controller logic, given a Reconciler.
type Base struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	// SharedClientSet allows us to configure shared objects
	SharedClientSet sharedclientset.Interface

	// KfClientSet allows us to configure Kf objects
	KfClientSet kfclientset.Interface

	// ConfigMapWatcher allows us to watch for ConfigMap changes.
	ConfigMapWatcher configmap.Watcher

	// Sugared logger is easier to use but is not as performant as the
	// raw logger. In performance critical paths, call logger.Desugar()
	// and use the returned raw logger instead. In addition to the
	// performance benefits, raw logger also preserves type-safety at
	// the expense of slightly greater verbosity.
	Logger *zap.SugaredLogger
}

// NewBase instantiates a new instance of Base implementing
// the common & boilerplate code between our reconcilers.
func NewBase(ctx context.Context, controllerAgentName string, cmw configmap.Watcher) *Base {
	logger := logging.FromContext(ctx).
		Named(controllerAgentName).
		With(zap.String(logkey.ControllerType, controllerAgentName))

	kubeClient := kubeclient.Get(ctx)

	base := &Base{
		KubeClientSet:    kubeClient,
		SharedClientSet:  sharedclient.Get(ctx),
		KfClientSet:      kfclient.Get(ctx),
		ConfigMapWatcher: cmw,
		Logger:           logger,
	}

	return base
}

func init() {
	// Add serving types to the default Kubernetes Scheme so Events can be
	// logged for serving types.
	kfscheme.AddToScheme(scheme.Scheme)
}
