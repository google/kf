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

package apiservercerts

import (
	"context"

	apiserviceclient "github.com/google/kf/v2/pkg/client/kube-aggregator/injection/client"
	apiserviceinformer "github.com/google/kf/v2/pkg/client/kube-aggregator/injection/informers/apiregistration/v1/apiservice"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/system"
	"k8s.io/client-go/tools/cache"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling API Service
// certs.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "upload.kf.dev")

	apiServiceInformer := apiserviceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	// Create the reconciler.
	r := &Reconciler{
		Base:                reconciler.NewBase(ctx, cmw),
		apiServiceLister:    apiServiceInformer.Lister(),
		apiServiceClientSet: apiserviceclient.Get(ctx),
	}

	impl := controller.NewContext(ctx, r, controller.ControllerOptions{WorkQueueName: "APIServices", Logger: logger})

	logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	apiServiceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(apiServiceName),
		Handler:    controller.HandleAll(impl.Enqueue),
	})
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithNameAndNamespace(system.Namespace(), SecretName),
		Handler:    controller.HandleAll(impl.Enqueue),
	})

	return impl
}
