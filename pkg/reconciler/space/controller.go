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

package space

import (
	"context"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	spaceinformer "github.com/google/kf/pkg/client/injection/informers/kf/v1alpha1/space"
	"github.com/google/kf/pkg/reconciler"
	namespaceinformer "github.com/knative/pkg/injection/informers/kubeinformers/corev1/namespace"
	roleinformer "github.com/knative/pkg/injection/informers/kubeinformers/rbacv1/role"

	"k8s.io/client-go/tools/cache"

	"github.com/knative/pkg/configmap"
	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/logging"
)

// NewController creates a new controller capable of reconciling Kf Spaces.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Get informers off context
	nsInformer := namespaceinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	roleInformer := roleinformer.Get(ctx)

	// Create reconciler
	c := &Reconciler{
		Base:            reconciler.NewBase(ctx, "space-controller", cmw),
		spaceLister:     spaceInformer.Lister(),
		namespaceLister: nsInformer.Lister(),
		roleLister:      roleInformer.Lister(),
	}

	impl := controller.NewImpl(c, logger, "Spaces")

	c.Logger.Info("Setting up event handlers")
	// Watch for changes in sub-resources so we can sync accordingly
	spaceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	nsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Space")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	roleInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Space")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
