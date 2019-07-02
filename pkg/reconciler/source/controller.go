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
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
)

// NewController creates a new controller capable of reconciling Kf sources.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Get informers off context
	sourceInformer := sourceinformer.Get(ctx)
	buildInformer := GetBuildInformer(ctx)

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatalf("failed to create a Build rest config: %s", err)
	}

	buildClient, err := cbuild.NewForConfig(config)
	if err != nil {
		logger.Fatalf("failed to create a Build client: %s", err)
	}

	// Create reconciler
	c := &Reconciler{
		Base:         reconciler.NewBase(ctx, "source-controller", cmw),
		SourceLister: sourceInformer.Lister(),
		buildLister:  buildInformer.Lister(),
		buildClient:  buildClient,
	}

	impl := controller.NewImpl(c, logger, "sources")

	c.Logger.Info("Setting up event handlers")

	// Watch for changes in sub-resources so we can sync accordingly
	sourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	buildInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kfv1alpha1.SchemeGroupVersion.WithKind("Source")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
