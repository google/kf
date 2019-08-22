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

package app

import (
	"context"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	appinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/app"
	routeinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/route"
	routeclaiminformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/routeclaim"
	sourceinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/source"
	spaceinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/space"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/injection/client"
	servicebindinginformer "github.com/google/kf/pkg/client/servicecatalog/injection/informers/servicecatalog/v1beta1/servicebinding"
	serviceinstanceinformer "github.com/google/kf/pkg/client/servicecatalog/injection/informers/servicecatalog/v1beta1/serviceinstance"
	"github.com/google/kf/pkg/reconciler"
	krevisioninformer "github.com/knative/serving/pkg/client/injection/informers/serving/v1alpha1/revision"
	kserviceinformer "github.com/knative/serving/pkg/client/injection/informers/serving/v1alpha1/service"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	secretinformer "knative.dev/pkg/injection/informers/kubeinformers/corev1/secret"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "apps.kf.dev")

	// Get informers off context
	knativeServiceInformer := kserviceinformer.Get(ctx)
	knativeRevisionInformer := krevisioninformer.Get(ctx)
	sourceInformer := sourceinformer.Get(ctx)
	appInformer := appinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	routeInformer := routeinformer.Get(ctx)
	routeClaimInformer := routeclaiminformer.Get(ctx)
	serviceBindingInformer := servicebindinginformer.Get(ctx)
	serviceInstanceInformer := serviceinstanceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	serviceCatalogClient := servicecatalogclient.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:                  reconciler.NewBase(ctx, cmw),
		serviceCatalogClient:  serviceCatalogClient,
		knativeServiceLister:  knativeServiceInformer.Lister(),
		knativeRevisionLister: knativeRevisionInformer.Lister(),
		sourceLister:          sourceInformer.Lister(),
		appLister:             appInformer.Lister(),
		secretLister:          secretInformer.Lister(),
		spaceLister:           spaceInformer.Lister(),
		routeLister:           routeInformer.Lister(),
		routeClaimLister:      routeClaimInformer.Lister(),
		serviceBindingLister:  serviceBindingInformer.Lister(),
		serviceInstanceLister: serviceInstanceInformer.Lister(),
	}

	impl := controller.NewImpl(c, logger, "Apps")

	logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	appInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	sourceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("App")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	knativeServiceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("App")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceBindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("App")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
