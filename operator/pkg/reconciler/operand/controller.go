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

package operand

import (
	"context"
	"sync"

	"kf-operator/pkg/apis/operand/v1alpha1"
	clientset "kf-operator/pkg/client/injection/client"
	clusteractiveoperandinformer "kf-operator/pkg/client/injection/informers/operand/v1alpha1/clusteractiveoperand"
	operandinformer "kf-operator/pkg/client/injection/informers/operand/v1alpha1/operand"
	operandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/operand"
	"kf-operator/pkg/operand/injection/dynamichelper"

	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const controllerAgentName = "operand-controller"

// NewController initializes the controller and is called by the generated code
// Registers eventhandlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	operandInformer := operandinformer.Get(ctx)
	clusteractiveoperandInformer := clusteractiveoperandinformer.Get(ctx)
	var lock sync.Mutex

	c := &reconciler{
		lock:                       &lock,
		operandClient:              clientset.Get(ctx).OperandV1alpha1(),
		clusterActiveOperandLister: clusteractiveoperandInformer.Lister(),
		resourceReconciler:         NewHealthcheckReconciler(NewManifestReconciler(dynamichelper.Get(ctx)), kubeclient.Get(ctx)),
	}
	impl := operandreconciler.NewImpl(ctx, c, func(_ *controller.Impl) controller.Options { return controller.Options{SkipStatusUpdates: true} })

	logger.Info("Setting up event handlers")
	operandInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	clusteractiveoperandInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("Operand").GroupKind()),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	c.enqueueAfter = impl.EnqueueAfter

	return impl
}
