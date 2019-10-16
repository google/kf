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
	"encoding/json"
	"fmt"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	routeinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/route"
	routeclaiminformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/routeclaim"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/reconciler"
	appresources "github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/google/kf/pkg/reconciler/route/config"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	virtualserviceinformer "knative.dev/pkg/client/injection/informers/istio/v1alpha3/virtualservice"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "routes.kf.dev")

	// Get informers off context
	vsInformer := virtualserviceinformer.Get(ctx)
	routeInformer := routeinformer.Get(ctx)
	routeClaimInformer := routeclaiminformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:                 reconciler.NewBase(ctx, cmw),
		routeLister:          routeInformer.Lister(),
		routeClaimLister:     routeClaimInformer.Lister(),
		virtualServiceLister: vsInformer.Lister(),
	}

	impl := controller.NewImpl(c, logger, "Routes")

	logger.Info("Setting up event handlers")

	enqueue := logError(logger.With("enqueue"), BuildEnqueuer(impl.Enqueue))

	routeInformer.Informer().AddEventHandler(
		controller.HandleAll(enqueue),
	)

	routeClaimInformer.Informer().AddEventHandler(
		controller.HandleAll(enqueue),
	)

	vsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: FilterVSWithNamespace(v1alpha1.KfNamespace),
		Handler:    controller.HandleAll(logError(logger, EnqueueRoutesOfVirtualService(enqueue, c.routeLister))),
	})

	// configmap stuff
	logger.Info("Setting up ConfigMap receivers")
	configsToResync := []interface{}{
		&config.RoutingConfig{},
	}
	resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		impl.GlobalResync(routeInformer.Informer())
	})
	configStore := config.NewStore(logger.Named("config-store"), controller.GetResyncPeriod(ctx), resync)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	return impl
}

// namespacedRouteSpecFields is used as a key for the route reconciler.
type namespacedRouteSpecFields struct {
	v1alpha1.RouteSpecFields
	Namespace string
}

// logError allows functions that assist with enqueing to return an error.
// Normal workflows work better when errors can be returned (instead of just
// logged). Therefore logError allows these functions to return an error and
// it will take care of swallowing and logging it.
func logError(logger *zap.SugaredLogger, f func(interface{}) error) func(interface{}) {
	return func(obj interface{}) {
		if err := f(obj); err != nil {
			logger.Warn(err)
		}
	}
}

// BuildEnqueuer returns a function that will enqueue a JSON marshalled
// namespacedRouteSpecFields from a Route or RouteClaim.
func BuildEnqueuer(enqueue func(interface{})) func(interface{}) error {
	return func(obj interface{}) error {
		nrf := namespacedRouteSpecFields{}
		switch r := obj.(type) {
		case *v1alpha1.Route:
			nrf = namespacedRouteSpecFields{
				Namespace:       r.GetNamespace(),
				RouteSpecFields: r.Spec.RouteSpecFields,
			}
		case *v1alpha1.RouteClaim:
			nrf = namespacedRouteSpecFields{
				Namespace:       r.GetNamespace(),
				RouteSpecFields: r.Spec.RouteSpecFields,
			}
		default:
			return fmt.Errorf("unexpected type: %T", obj)
		}

		data, err := json.Marshal(nrf)
		if err != nil {
			// This should never happen
			return err
		}

		enqueue(cache.ExplicitKey(data))
		return nil
	}
}

// FilterVSWithNamespace makes it simple to create FilterFunc's for use with
// cache.FilteringResourceEventHandler that filter based on a namespace and if
// the type is a VirtualService.
func FilterVSWithNamespace(namespace string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		if object, ok := obj.(metav1.Object); ok {
			if namespace == object.GetNamespace() {
				_, ok := obj.(*networking.VirtualService)
				return ok
			}
		}
		return false
	}
}

// EnqueueRoutesOfVirtualService will find the corresponding routes for the
// VirtualService.  It will Enqueue a key for each one. We aren't able to use
// EnqueueControllerOf (as other components do), because a VirtualService is
// NOT owned by a single Route. Therefore, when one changes, we need to grab
// the collection of corresponding Routes.
func EnqueueRoutesOfVirtualService(
	enqueue func(interface{}),
	routeLister kflisters.RouteLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		vs, ok := obj.(*networking.VirtualService)
		if !ok {
			return nil
		}

		routes, err := routeLister.
			Routes(vs.Annotations["space"]).
			List(appresources.MakeRouteSelectorNoPath(v1alpha1.RouteSpecFields{
				Domain:   vs.Annotations["domain"],
				Hostname: vs.Annotations["hostname"],
			}))
		if err != nil {
			return fmt.Errorf("failed to list corresponding routes: %s", err)
		}

		for _, route := range routes {
			enqueue(route)
		}

		return nil
	}
}
