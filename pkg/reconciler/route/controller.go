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

package route

import (
	"context"
	"time"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	routeinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/route"
	"github.com/google/kf/pkg/reconciler"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	kserviceinformer "github.com/knative/serving/pkg/client/injection/informers/serving/v1alpha1/service"
	virtualserviceinformer "knative.dev/pkg/client/injection/informers/istio/v1alpha3/virtualservice"

	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Get informers off context
	vsInformer := virtualserviceinformer.Get(ctx)
	routeInformer := routeinformer.Get(ctx)
	serviceInformer := kserviceinformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:                 reconciler.NewBase(ctx, "route-controller", cmw),
		routeLister:          routeInformer.Lister(),
		virtualServiceLister: vsInformer.Lister(),
	}

	impl := controller.NewImpl(c, logger, "Routes")

	c.Logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	routeInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	vsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Route")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			start := time.Now()
			service := obj.(*serving.Service)
			namespaceAndName := service.GetNamespace() + "/" + service.GetName()

			c.Logger.Infof("Deleting references to service %s in routes", namespaceAndName)

			if err := c.ReconcileServiceDeletion(ctx, service); err != nil {
				c.Logger.Warnf("failed to delete references to service %s in routes: %s", namespaceAndName, err)
			}
			c.Logger.Infof("Reconcile (service deletion) succeeded. Time taken: %s.", time.Since(start))
		},
	})

	return impl
}
