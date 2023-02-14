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

package task

import (
	"context"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"

	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	taskinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/task"
	"github.com/google/kf/v2/pkg/reconciler"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf Tasks.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "tasks.kf.dev")

	// Get informers off context.
	appInformer := appinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	taskInformer := taskinformer.Get(ctx)
	taskRunInformer := taskruninformer.Get(ctx)
	tektonClient := tektonclient.Get(ctx)

	taskLister := taskInformer.Lister()

	// Create reconciler.
	c := &Reconciler{
		Base:          reconciler.NewBase(ctx, cmw),
		appLister:     appInformer.Lister(),
		spaceLister:   spaceInformer.Lister(),
		taskLister:    taskLister,
		taskRunLister: taskRunInformer.Lister(),
		tektonClient:  tektonClient.TektonV1beta1(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "tasks",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("Setting up ConfigMap receivers")
	configsToResync := []interface{}{
		&config.DefaultsConfig{},
	}
	resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		// don't cause a resync on update because tasks are only executed
		// once.
	})
	configStore := config.NewStore(logger.Named("kf-config-store"), resync)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	logger.Info("Setting up event handlers")

	taskInformer.Informer().AddEventHandlerWithResyncPeriod(
		controller.HandleAll(impl.Enqueue),
		5*time.Minute,
	)

	logger.Info("Setting up task run event handlers")

	taskRunInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("Task")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
