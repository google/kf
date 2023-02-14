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

package serviceinstance

import (
	"context"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfserviceinstanceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstance"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"k8s.io/client-go/tools/cache"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	persistentvolumeinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/persistentvolume"
	persistentvolumeclaiminformer "knative.dev/pkg/client/injection/kube/informers/core/v1/persistentvolumeclaim"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Routes.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "serviceinstances.kf.dev")

	// Get informers off context
	serviceInstanceInformer := kfserviceinstanceinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	k8sServiceInformer := serviceinformer.Get(ctx)

	// These informers are not used to watch events.
	// PVs and PVCs are immutable. The reconciler only creates them if they don't exist.
	persistentVolumeInformer := persistentvolumeinformer.Get(ctx)
	persistentVolumeClaimInformer := persistentvolumeclaiminformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		ServiceCatalogBase: reconciler.NewServiceCatalogBase(ctx, cmw),
		spaceLister:        spaceInformer.Lister(),
		deploymentLister:   deploymentInformer.Lister(),
		volumeLister:       persistentVolumeInformer.Lister(),
		volumeClaimLister:  persistentVolumeClaimInformer.Lister(),
		k8sServiceLister:   k8sServiceInformer.Lister(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "ServiceInstances",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("Setting up event handlers")

	// Resync all service instances at least once a minute so we can
	// poll async operations.
	serviceInstanceInformer.Informer().AddEventHandlerWithResyncPeriod(
		controller.HandleAll(impl.Enqueue),
		1*time.Minute,
	)

	// Set up all owned resources to be triggered only based on the controller.
	for _, informer := range []cache.SharedIndexInformer{
		secretInformer.Informer(),
		deploymentInformer.Informer(),
		k8sServiceInformer.Informer(),
	} {
		informer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ServiceInstance")),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})
	}

	logger.Info("Setting up ConfigMap receivers")
	configsToResync := []interface{}{
		&config.DefaultsConfig{},
	}
	resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		impl.GlobalResync(spaceInformer.Informer())
	})
	configStore := config.NewStore(logger.Named("kf-config-store"), resync)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	return impl
}
