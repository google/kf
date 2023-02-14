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

package clusterservicebroker

import (
	"context"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclusterservicebrokerinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/clusterservicebroker"
	kfserviceinstanceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstance"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"k8s.io/client-go/tools/cache"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "clusterservicebrokers.kf.dev")

	// Get informers off context
	clusterServiceBrokerInformer := kfclusterservicebrokerinformer.Get(ctx)
	serviceInstanceInformer := kfserviceinstanceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	r := &Reconciler{
		Base:                         reconciler.NewBase(ctx, cmw),
		kfClusterServiceBrokerLister: clusterServiceBrokerInformer.Lister(),
		kfServiceInstanceLister:      serviceInstanceInformer.Lister(),
	}

	impl := controller.NewContext(ctx, r, controller.ControllerOptions{
		WorkQueueName: "ClusterServiceBrokers",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("setting up event handlers")
	clusterServiceBrokerInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Set up all owned resources to be triggered only based on the controller.
	ownedResources := []cache.SharedIndexInformer{
		secretInformer.Informer(),
	}

	for _, informer := range ownedResources {
		informer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ClusterServiceBroker")),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})
	}

	return impl
}
