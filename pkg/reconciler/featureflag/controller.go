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

package featureflag

import (
	"context"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"k8s.io/client-go/tools/cache"
	namespaceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/namespace"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "featureflag")

	namespaceInformer := namespaceinformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:            reconciler.NewBase(ctx, cmw),
		namespaceLister: namespaceInformer.Lister(),
	}

	logger.Info("Setting up event handlers")

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "FeatureFlag",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	// Run the reconciler every 5 minutes to update the status of feature flags (e.g., route services).
	namespaceInformer.Informer().AddEventHandlerWithResyncPeriod(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(v1alpha1.KfNamespace),
		Handler:    controller.HandleAll(impl.Enqueue),
	}, 5*time.Minute)

	logger.Info("Setting up ConfigMap receivers")
	configsToResync := []interface{}{
		&config.DefaultsConfig{},
	}
	resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		impl.FilteredGlobalResync(
			controller.FilterWithName(v1alpha1.KfNamespace),
			namespaceInformer.Informer())
	})
	configStore := config.NewStore(logger.Named("kf-config-store"), resync)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	return impl
}
