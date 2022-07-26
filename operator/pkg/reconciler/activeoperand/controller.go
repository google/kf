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

package activeoperand

import (
	"context"

	activeoperandinformer "kf-operator/pkg/client/injection/informers/operand/v1alpha1/activeoperand"
	activeoperandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/activeoperand"
	"kf-operator/pkg/operand"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const controllerAgentName = "activeoperand-controller"

// NewController initializes the controller and is called by the generated code
// Registers eventhandlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	activeoperandInformer := activeoperandinformer.Get(ctx)

	c := &reconciler{OwnerHandler: operand.CreateOwnerHandlerWithCtx(ctx)}

	impl := activeoperandreconciler.NewImpl(ctx, c)

	logger.Info("Setting up event handlers")
	activeoperandInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	c.enqueueAfter = impl.EnqueueAfter
	return impl
}
