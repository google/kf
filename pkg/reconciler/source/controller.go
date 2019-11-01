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

package source

import (
	"context"

	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	sourceinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/source"
	"github.com/google/kf/pkg/reconciler"
	taskrunclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/taskrun"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
)

// NewController creates a new controller capable of reconciling Kf sources.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "sources.kf.dev")

	// Get informers off context
	sourceInformer := sourceinformer.Get(ctx)
	taskRunInformer := taskruninformer.Get(ctx)
	taskRunClient := taskrunclient.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:          reconciler.NewBase(ctx, cmw),
		sourceLister:  sourceInformer.Lister(),
		taskRunLister: taskRunInformer.Lister(),
		taskRunClient: taskRunClient.TektonV1alpha1(),
	}

	impl := controller.NewImpl(c, logger, "sources")

	logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	sourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	taskRunInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("Source")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	c.SecretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("Source")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
