// Copyright 2020 Google LLC
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

package serviceinstancebinding

import (
	"context"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	serviceinstanceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstance"
	kfservicebindinginformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstancebinding"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "serviceinstancebindings.kf.dev")

	// Get informers off context
	appInformer := appinformer.Get(ctx)
	serviceBindingInformer := kfservicebindinginformer.Get(ctx)
	serviceInstanceInformer := serviceinstanceinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		ServiceCatalogBase: reconciler.NewServiceCatalogBase(ctx, cmw),
		spaceLister:        spaceInformer.Lister(),
		appLister:          appInformer.Lister(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "ServiceInstanceBindings",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	serviceBindingInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Watch for changes in ServiceInstances to reconcile bindings when
	// ServiceInstances become Ready.
	serviceInstanceInformer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		UpdateFunc: controller.PassNew(
			reconciler.LogEnqueueError(logger,
				enqueueBindingsForService(impl.Enqueue, serviceBindingInformer.Lister()))),
		DeleteFunc: nil,
	})

	// Watch for changes in Secrets that are owned by Service Instances so we
	// can propagate the changes to credentials.
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ServiceInstance")),
		Handler: controller.HandleAll(reconciler.LogEnqueueError(
			logger,
			enqueueServiceInstanceSecrets(impl.Enqueue, serviceBindingInformer.Lister()),
		)),
	})

	// Set up all owned resources to be triggered only based on the controller.
	for _, informer := range []cache.SharedIndexInformer{
		secretInformer.Informer(),
	} {
		informer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ServiceInstanceBinding")),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})
	}

	return impl
}

// enqueueBindingsForService enqueues all ServiceInstanceBindings for the given ServiceInstance.
func enqueueBindingsForService(
	enqueue func(interface{}),
	serviceBindingLister kflisters.ServiceInstanceBindingLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		service, ok := obj.(*v1alpha1.ServiceInstance)
		if !ok {
			return nil
		}

		bindings, err := serviceBindingLister.
			ServiceInstanceBindings(service.Namespace).
			List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed to list bindings: %s", err)
		}

		for _, binding := range bindings {
			if binding.Spec.InstanceRef.Name == service.Name {
				enqueue(binding)
			}
		}
		return nil
	}
}

func enqueueServiceInstanceSecrets(
	enqueue func(interface{}),
	serviceBindingLister kflisters.ServiceInstanceBindingLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			// This should not happen because due to the filter...
			return nil
		}

		serviceInstance := metav1.GetControllerOfNoCopy(secret)
		if serviceInstance == nil {
			// This should not happen because due to the filter...
			return nil
		}

		bindings, err := serviceBindingLister.
			ServiceInstanceBindings(secret.GetNamespace()).
			List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed to get ServiceInstance Bindings (%s): %v", secret.GetNamespace(), err)
		}

		for _, binding := range bindings {
			if binding.Spec.InstanceRef.Name != serviceInstance.Name {
				continue
			}

			// Found the corresponding Binding for the Service Instance.
			enqueue(binding)
			return nil
		}

		return nil
	}
}
