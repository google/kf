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
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	routeinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/route"
	serviceinstancebindinginformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstancebinding"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	networkingclient "github.com/google/kf/v2/pkg/client/networking/injection/client"
	virtualserviceinformer "github.com/google/kf/v2/pkg/client/networking/injection/informers/networking/v1alpha3/virtualservice"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/route/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "routes.kf.dev")

	// Get informers off context
	vsInformer := virtualserviceinformer.Get(ctx)
	routeInformer := routeinformer.Get(ctx)
	appInformer := appinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	serviceInstanceBindingInformer := serviceinstancebindinginformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:                         reconciler.NewBase(ctx, cmw),
		appLister:                    appInformer.Lister(),
		spaceLister:                  spaceInformer.Lister(),
		routeLister:                  routeInformer.Lister(),
		virtualServiceLister:         vsInformer.Lister(),
		networkingClientSet:          networkingclient.Get(ctx),
		serviceInstanceBindingLister: serviceInstanceBindingInformer.Lister(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{WorkQueueName: "Routes", Logger: logger})

	logger.Info("Setting up event handlers")

	enqueue := reconciler.LogEnqueueError(logger, BuildEnqueuer(impl.EnqueueKey))

	appInformer.Informer().AddEventHandler(
		controller.HandleAll(enqueue),
	)

	routeInformer.Informer().AddEventHandler(
		controller.HandleAll(enqueue),
	)

	vsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: FilterVSManagedByKf(),
		Handler:    controller.HandleAll(EnqueueRoutesOfVirtualService(enqueue)),
	})

	serviceInstanceBindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		// Accept all service instance bindings that bind a service to a route
		FilterFunc: func(obj interface{}) bool {
			sb, ok := obj.(*v1alpha1.ServiceInstanceBinding)
			if !ok {
				logger.Error("failed to cast obj to service instance binding")
				return false
			}
			return sb.Spec.Route != nil
		},
		// Enqueue route in service instance binding
		Handler: controller.HandleAll(enqueue),
	})

	return impl
}

// BuildEnqueuer returns a function that will enqueue a JSON marshalled
// RouteSpecFields from a Route or Route.
func BuildEnqueuer(enqueue func(types.NamespacedName)) func(interface{}) error {
	return func(obj interface{}) error {

		switch r := obj.(type) {
		case *v1alpha1.App:
			// Add domains regardless of orphaned so the owners holding on to the
			// orphans get reconciled and release their hold of the routes.
			for _, rt := range r.Status.Routes {
				enqueue(types.NamespacedName{
					Namespace: r.GetNamespace(),
					Name:      rt.Source.Domain,
				})
			}
		case *v1alpha1.Route:
			enqueue(types.NamespacedName{
				Namespace: r.GetNamespace(),
				Name:      r.Spec.RouteSpecFields.Domain,
			})
		case *networking.VirtualService:
			if domain, ok := r.Annotations[resources.DomainAnnotation]; ok {
				enqueue(types.NamespacedName{
					Namespace: r.GetNamespace(),
					Name:      domain,
				})
			}
		case *v1alpha1.ServiceInstanceBinding:
			routeSpecFields := r.Spec.Route
			if routeSpecFields != nil {
				enqueue(types.NamespacedName{
					Namespace: r.GetNamespace(),
					Name:      routeSpecFields.Domain,
				})
			}
		default:
			return fmt.Errorf("unexpected type: %T", obj)
		}

		return nil
	}
}

// FilterVSManagedByKf makes it simple to create FilterFunc's for use with
// cache.FilteringResourceEventHandler that filter based on the
// "app.kubernetes.io/managed-by": "kf" label and if the type is a VirtualService.
func FilterVSManagedByKf() func(obj interface{}) bool {
	return func(obj interface{}) bool {
		if object, ok := obj.(metav1.Object); ok {
			if "kf" == object.GetLabels()[v1alpha1.ManagedByLabel] {
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
func EnqueueRoutesOfVirtualService(enqueue func(interface{})) func(obj interface{}) {
	return func(obj interface{}) {
		vs, ok := obj.(*networking.VirtualService)
		if !ok {
			return
		}

		if _, ok := vs.Annotations[resources.DomainAnnotation]; ok {
			enqueue(vs)
		}
		return
	}
}
