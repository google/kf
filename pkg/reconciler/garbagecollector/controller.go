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

package garbagecollector

import (
	"context"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	buildinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/build"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

const GarbageCollectionInterval time.Duration = 5 * time.Minute

// NewController creates a new controller capable of reconciling Kf Apps.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "garbagecollector.kf.dev")

	// Get informers off context
	buildInformer := buildinformer.Get(ctx)
	appInformer := appinformer.Get(ctx)
	appLister := appInformer.Lister()

	// Create reconciler
	c := &Reconciler{
		Base:        reconciler.NewBase(ctx, cmw),
		buildLister: buildInformer.Lister(),
		appLister:   appLister,
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "GarbageCollector",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
		Concurrency:   10,
	})

	logger.Info("Setting up ConfigMap receivers")

	kfConfigStore := kfconfig.NewStore(logger.Named("kf-config-store"))
	kfConfigStore.WatchConfigs(cmw)
	c.kfConfigStore = kfConfigStore

	// Watch for changes in sub-resources so we can sync accordingly
	appInformer.Informer().AddEventHandlerWithResyncPeriod(controller.HandleAll(impl.Enqueue), GarbageCollectionInterval)

	return impl
}
