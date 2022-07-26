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
	"flag"
	"fmt"
	"time"

	"kf-operator/pkg/apis/operand/v1alpha1"
	activeoperandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/activeoperand"
	"kf-operator/pkg/operand"

	"knative.dev/pkg/kmeta"

	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for ActiveOperand resources.
type reconciler struct {
	operand.OwnerHandler
	enqueueAfter func(interface{}, time.Duration)
}

var (
	reconcilePeriod time.Duration
)

func init() {
	flag.DurationVar(&reconcilePeriod, "ao_reconcile_period", time.Minute, "Period with which to reconcile the AO CR when there are no changes to it. Defaults to 1 minute")
}

// Check that our Reconciler implements activeoperandreconciler.Interface
var _ activeoperandreconciler.Interface = (*reconciler)(nil)

func (r reconciler) ReconcileKind(ctx context.Context, ao *v1alpha1.ActiveOperand) pkgreconciler.Event {
	defer r.enqueueAfter(ao, reconcilePeriod)
	logger := logging.FromContext(ctx)
	ao.Status.InitializeConditions()
	ao.Status.ObservedGeneration = ao.GetGeneration()

	logger.Info("Injecting namespace ownerrefs")
	err := r.OwnerHandler.HandleOwnerRefs(ctx, kmeta.NewControllerRef(ao), ao.Spec.Live, ao)
	if err != nil {
		logger.Info("Inject namespace ownerrefs failed.")
		ao.Status.MarkOwnerRefsInjectedFailed(fmt.Sprintf("%+v", err))
	} else {
		logger.Info("Inject namespace ownerrefs success.")
		ao.Status.MarkOwnerRefsInjected()
	}
	return err
}
