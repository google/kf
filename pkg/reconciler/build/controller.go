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

package build

import (
	"context"
	"fmt"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	buildinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/build"
	sourcepackageinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/sourcepackage"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf builds.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "builds.kf.dev")

	// Get informers off context
	buildInformer := buildinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	taskRunInformer := taskruninformer.Get(ctx)
	sourcePackageInformer := sourcepackageinformer.Get(ctx)
	tektonClient := tektonclient.Get(ctx)

	buildLister := buildInformer.Lister()

	// Create reconciler
	c := &Reconciler{
		Base:                reconciler.NewBase(ctx, cmw),
		buildLister:         buildLister,
		spaceLister:         spaceInformer.Lister(),
		sourcePackageLister: sourcePackageInformer.Lister(),
		taskRunLister:       taskRunInformer.Lister(),
		tektonClient:        tektonClient.TektonV1beta1(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "Builds",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("Setting up ConfigMap receivers")

	configStore := config.NewStore(logger.Named("secrets-config-store"))
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	kfConfigStore := kfconfig.NewStore(logger.Named("kf-config-store"))
	kfConfigStore.WatchConfigs(cmw)
	c.kfConfigStore = kfConfigStore

	logger.Info("Setting up event handlers")

	// Watch for space changes and enqueue all builds in the evented space
	spaceInformer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: controller.PassNew(reconciler.LogEnqueueError(logger, EnqueueBuildsOfSpace(impl.Enqueue, buildLister))),
		DeleteFunc: nil,
	})

	// Watch for changes in sub-resources so we can sync accordingly
	buildInformer.Informer().AddEventHandlerWithResyncPeriod(
		controller.HandleAll(impl.Enqueue),
		5*time.Minute,
	)

	taskRunInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("Build")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// SourcePackages aren't owned by a Build, but we still want to trigger
	// off of them.
	sourcePackageInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		// NOTE: Source packages are owned by the App (not the Build).
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("App")),
		Handler: controller.HandleAll(
			reconciler.LogEnqueueError(logger,
				EnqueueBuildsOfSourcePackage(
					impl.Enqueue,
					buildLister,
				),
			),
		),
	})

	return impl
}

// EnqueueBuildsOfSpace will find the corresponding Builds for the
// Space. It will enqueue a key for each one.
func EnqueueBuildsOfSpace(
	enqueue func(interface{}),
	buildLister kflisters.BuildLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		space, ok := obj.(*v1alpha1.Space)
		if !ok {
			return nil
		}

		builds, err := buildLister.
			Builds(space.Name).
			List(labels.Everything())

		if err != nil {
			return fmt.Errorf("failed to list corresponding Builds: %s", err)
		}

		for _, build := range builds {
			enqueue(build)
		}

		return nil
	}
}

// EnqueueBuildsOfSourcePackage will find the corresponding Builds for the
// SourcePackage. It will enqueue the key for any it finds.
func EnqueueBuildsOfSourcePackage(
	enqueue func(interface{}),
	buildLister kflisters.BuildLister,
) func(obj interface{}) error {
	return func(obj interface{}) error {
		sourcePackage, ok := obj.(*v1alpha1.SourcePackage)
		if !ok {
			return nil
		}

		builds, err := buildLister.
			Builds(sourcePackage.Namespace).
			List(labels.Everything())

		if err != nil {
			return fmt.Errorf("failed to list corresponding Builds: %s", err)
		}

		for _, build := range builds {
			if build.Spec.SourcePackage.Name != sourcePackage.Name {
				continue
			}
			enqueue(build)
		}

		return nil
	}
}
