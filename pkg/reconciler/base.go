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
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"

	sharedclient "github.com/knative/pkg/client/injection/client"
	"github.com/knative/pkg/injection/clients/kubeclient"
	"github.com/poy/service-catalog/pkg/client/clientset_generated/clientset"

	sharedclientset "github.com/knative/pkg/client/clientset/versioned"
	"github.com/knative/pkg/configmap"
)

// Base implements the core controller logic, given a Reconciler.
type Base struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	// SharedClientSet allows us to configure shared objects
	SharedClientSet sharedclientset.Interface

	// KfClientSet allows us to configure Kf objects
	KfClientSet clientset.Interface

	// ConfigMapWatcher allows us to watch for ConfigMap changes.
	ConfigMapWatcher configmap.Watcher

	Logger *log.Logger
}

// NewBase instantiates a new instance of Base implementing
// the common & boilerplate code between our reconcilers.
func NewBase(ctx context.Context, controllerAgentName string, cmw configmap.Watcher) *Base {
	kubeClient := kubeclient.Get(ctx)

	base := &Base{
		KubeClientSet:    kubeClient,
		SharedClientSet:  sharedclient.Get(ctx),
		ConfigMapWatcher: cmw,
		Logger:           log.New(os.Stdout, fmt.Sprintf("%s > ", controllerAgentName), 0),
	}

	return base
}
