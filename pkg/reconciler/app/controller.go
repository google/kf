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
	"fmt"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclientset "github.com/google/kf/v2/pkg/client/kf/clientset/versioned"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	buildinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/build"
	routeinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/route"
	serviceinstancebindinginformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstancebinding"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	autoscalinginformer "knative.dev/pkg/client/injection/kube/informers/autoscaling/v1/horizontalpodautoscaler"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	serviceaccountinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

// NewController creates a new controller capable of reconciling Kf Apps.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "apps.kf.dev")

	// Get informers off context
	buildInformer := buildinformer.Get(ctx)
	appInformer := appinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	routeInformer := routeinformer.Get(ctx)
	serviceInstanceBindingInformer := serviceinstancebindinginformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	serviceAccountInformer := serviceaccountinformer.Get(ctx)
	hpaInformer := autoscalinginformer.Get(ctx)

	appLister := appInformer.Lister()

	// Dynamic client.
	dynamicClient := dynamicclient.Get(ctx)
	dynamicInformer := dynamicinformer.NewDynamicSharedInformerFactory(
		dynamicClient,
		time.Minute,
	)

	buildsGVR := schema.GroupVersionResource{
		Group:    "builds.appdevexperience.dev",
		Version:  "v1alpha1",
		Resource: "builds",
	}
	adxBuildInformer := dynamicInformer.ForResource(buildsGVR)

	// Create reconciler
	c := &Reconciler{
		Base:                         reconciler.NewBase(ctx, cmw),
		buildLister:                  buildInformer.Lister(),
		appLister:                    appLister,
		spaceLister:                  spaceInformer.Lister(),
		routeLister:                  routeInformer.Lister(),
		serviceInstanceBindingLister: serviceInstanceBindingInformer.Lister(),
		deploymentLister:             deploymentInformer.Lister(),
		serviceLister:                serviceInformer.Lister(),
		serviceAccountLister:         serviceAccountInformer.Lister(),
		autoscalingLister:            hpaInformer.Lister(),
		adxBuildLister:               adxBuildInformer.Lister(),
	}

	// We only want to start this informer if the ADX build type is installed.
	if isADXBuildInstalled(ctx, c.KfClientSet, logger) {
		go adxBuildInformer.Informer().Run(ctx.Done())
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "Apps",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
		Concurrency:   10,
	})

	logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	appInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Watch for Space changes and enqueue all Apps in the evented Space
	spaceInformer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: controller.PassNew(reconciler.LogEnqueueError(logger, enqueueAppsOfSpace(impl.Enqueue, appLister))),
		DeleteFunc: nil,
	})

	logger.Info("Setting up ConfigMap receivers")

	kfConfigStore := kfconfig.NewStore(logger.Named("kf-config-store"))
	kfConfigStore.WatchConfigs(cmw)
	c.kfConfigStore = kfConfigStore

	// Set up all owned resources to be triggered only based on the controller.
	for _, informer := range []cache.SharedIndexInformer{
		buildInformer.Informer(),
		adxBuildInformer.Informer(),
		deploymentInformer.Informer(),
		serviceInformer.Informer(),
		serviceAccountInformer.Informer(),
		hpaInformer.Informer(),
	} {
		informer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("App")),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})
	}

	c.SecretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ServiceInstanceBinding")),
		Handler: controller.HandleAll(reconciler.LogEnqueueError(
			logger,
			enqueueServiceInstanceBindingSecrets(
				impl.Enqueue,
				serviceInstanceBindingInformer.Lister(),
				appInformer.Lister(),
			)),
		),
	})

	serviceInstanceBindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		// Accept all service instance bindings that bind a service to an app
		FilterFunc: func(obj interface{}) bool {
			sb, ok := obj.(*v1alpha1.ServiceInstanceBinding)
			if !ok {
				logger.Error("failed to cast obj to service instance binding")
				return false
			}
			return sb.Spec.App != nil
		},
		// Enqueue app in service instance binding
		Handler: controller.HandleAll(func(obj interface{}) {
			sb, ok := obj.(*v1alpha1.ServiceInstanceBinding)
			if !ok {
				logger.Error("failed to cast obj service instance binding")
				return
			}
			impl.EnqueueKey(types.NamespacedName{
				Namespace: sb.GetNamespace(),
				Name:      sb.Spec.App.Name,
			})
		}),
	})

	routeInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			return true
		},
		// Enqueue apps whose Routes have changed.
		Handler: controller.HandleAll(func(obj interface{}) {
			rc, ok := obj.(*v1alpha1.Route)
			if !ok {
				logger.Error("failed to cast obj to a Route")
				return
			}

			apps, err := appLister.Apps(rc.Namespace).List(labels.Everything())
			if err != nil {
				logger.Errorf("failed to list apps: %v", err)
				return
			}
			for _, app := range apps {
				for _, routes := range app.Status.Routes {
					if rc.Spec.Domain == routes.Source.Domain {
						impl.EnqueueKey(types.NamespacedName{
							Namespace: app.Namespace,
							Name:      app.Name,
						})
						break
					}
				}
			}
		}),
	})

	return impl
}

func enqueueServiceInstanceBindingSecrets(
	enqueue func(interface{}),
	lister kflisters.ServiceInstanceBindingLister,
	appLister kflisters.AppLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			// This should not happen because due to the filter...
			return nil
		}

		// Find the owning ServiceInstance.
		owner := metav1.GetControllerOfNoCopy(secret)
		if owner == nil {
			// This should not happen because due to the filter...
			return nil
		}

		binding, err := lister.
			ServiceInstanceBindings(secret.GetNamespace()).
			Get(owner.Name)
		if err != nil {
			return fmt.Errorf("failed to get ServiceInstance Bindings (%s/%s): %v", secret.GetNamespace(), owner.Name, err)
		}

		if binding.Spec.App == nil {
			// We're only interested in App bindings.
			return nil
		}

		// Found the corresponding binding. Now find the App.
		appName := binding.Spec.App.Name
		app, err := appLister.
			Apps(secret.GetNamespace()).
			Get(appName)
		if err != nil {
			return fmt.Errorf("failed to find App (%s) from Service Instance Binding (%s): %v", appName, binding.Name, err)
		}

		// Enqueue the App.
		enqueue(app)
		return nil
	}
}

// enqueueAppsOfSpace will find the corresponding Apps for the Space. It will
// Enqueue a key for each one.
func enqueueAppsOfSpace(
	enqueue func(interface{}),
	appLister kflisters.AppLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		space, ok := obj.(*v1alpha1.Space)
		if !ok {
			return nil
		}

		apps, err := appLister.
			Apps(space.Name).
			List(labels.Everything())

		if err != nil {
			return fmt.Errorf("failed to list corresponding apps: %s", err)
		}

		for _, app := range apps {
			enqueue(app)
		}

		return nil
	}
}

func isADXBuildInstalled(
	ctx context.Context,
	client kfclientset.Interface,
	logger *zap.SugaredLogger,
) bool {
	list, err := client.
		Discovery().
		ServerResourcesForGroupVersion("builds.appdevexperience.dev/v1alpha1")
	if err != nil {
		logger.Warnf("failed to fetch resources: %v", err)
		return false
	}

	for _, r := range list.APIResources {
		if r.Kind == "Build" {
			return true
		}
	}

	return false
}
