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

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	routeinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/route"
	"github.com/google/kf/pkg/reconciler"
	appresources "github.com/google/kf/pkg/reconciler/app/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	virtualserviceinformer "knative.dev/pkg/client/injection/informers/istio/v1alpha3/virtualservice"
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

	// Watch for any changes to VirtualServices in the kf namespace.
	vsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: FilterVSWithNamespace(v1alpha1.KfNamespace),
		Handler:    controller.HandleAll(EnqueueRoutesOfVirtualService(impl, c)),
	})

	return impl
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
	c *controller.Impl,
	r *Reconciler,
) func(obj interface{}) {
	return func(obj interface{}) {
		vs, ok := obj.(*networking.VirtualService)
		if !ok {
			return
		}

		routes, err := r.routeLister.
			Routes(vs.Annotations["space"]).
			List(appresources.MakeRouteSelectorNoPath(v1alpha1.RouteSpecFields{
				Domain:   vs.Annotations["domain"],
				Hostname: vs.Annotations["hostname"],
			}))
		if err != nil {
			r.Logger.Warnf("failed to list corresponding routes: %s", err)
			return
		}

		for _, route := range routes {
			c.Enqueue(route)
		}
	}
}
