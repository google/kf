/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kfsystem

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"k8s.io/client-go/kubernetes"

	"kf-operator/pkg/apis/kfsystem/v1alpha1"
	clientset "kf-operator/pkg/client/clientset/versioned"
	kfreconciler "kf-operator/pkg/client/injection/reconciler/kfsystem/v1alpha1/kfsystem"
	kfoperand "kf-operator/pkg/operand/kf"

	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for KfSystem resources.
type Reconciler struct {
	kubeClient  kubernetes.Interface
	client      clientset.Interface
	reconcilers []kfoperand.Interface
}

// Check that our Reconciler implements cloudrunreconciler.Interface
var _ kfreconciler.Interface = (*Reconciler)(nil)
var _ kfreconciler.Finalizer = (*Reconciler)(nil)

var (
	reconcilePeriod       time.Duration
	reconcileFailedPeriod time.Duration
)

// ReconcileKind is called on the creation or mutation of a KfSystem CR.
func (r *Reconciler) ReconcileKind(ctx context.Context, kf *v1alpha1.KfSystem) pkgreconciler.Event {
	// Always acknowledge we saw the thing.
	kf.Status.InitializeConditions()
	// We allow other reconcilers to see the generation change before modifying
	defer func() { kf.Status.ObservedGeneration = kf.GetGeneration() }()
	logger := logging.FromContext(ctx)
	logger.Infof("Got KfSystem CR, reconciling %+v", kf)

	var result multierror.Group
	for _, reconciler := range r.reconcilers {
		// Closure captures the reconciler in each iteration.
		reconciler := reconciler
		result.Go(func() error { return reconciler.Reconcile(ctx, kf) })
	}
	return result.Wait().ErrorOrNil()
}

// FinalizeKind finalizes v1alpha1.KfSystem.
func (r *Reconciler) FinalizeKind(ctx context.Context, kfs *v1alpha1.KfSystem) pkgreconciler.Event {
	kfs.Status.ObservedGeneration = kfs.GetGeneration()
	logger := logging.FromContext(ctx)
	logger.Infof("Got KfSystem CR, finalizing %+v", kfs)

	var result multierror.Group
	for _, reconciler := range r.reconcilers {
		// Closure captures the reconciler in each iteration.
		reconciler := reconciler
		result.Go(func() error { return reconciler.Finalize(ctx) })
	}
	return result.Wait().ErrorOrNil()
}
