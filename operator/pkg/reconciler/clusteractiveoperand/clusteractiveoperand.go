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

package clusteractiveoperand

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"kf-operator/pkg/apis/operand/v1alpha1"
	opclient "kf-operator/pkg/client/clientset/versioned/typed/operand/v1alpha1"
	clusteractiveoperandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/clusteractiveoperand"
	"kf-operator/pkg/operand"

	"knative.dev/pkg/ptr"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for ClusterActiveOperand resources.
type clusterReconciler struct {
	operand.OwnerHandler
	operandGetter opclient.OperandV1alpha1Interface
	enqueueAfter  func(interface{}, time.Duration)
}

var (
	reconcilePeriod time.Duration
)

func init() {
	flag.DurationVar(&reconcilePeriod, "cao_reconcile_period", time.Minute, "Period with which to reconcile the CAO CR when there are no changes to it. Defaults to 1 minute")
}

func deleteOptions() metav1.DeleteOptions {
	do := &metav1.DeleteOptions{GracePeriodSeconds: ptr.Int64(30)}
	policy := metav1.DeletePropagationForeground
	do.PropagationPolicy = &policy
	return *do
}

// Check that our Reconciler implements clusteractiveoperandclusterReconciler.Interface
var _ clusteractiveoperandreconciler.Interface = (*clusterReconciler)(nil)

func (r clusterReconciler) ReconcileKind(ctx context.Context, ao *v1alpha1.ClusterActiveOperand) pkgreconciler.Event {
	defer r.enqueueAfter(ao, reconcilePeriod)
	logger := logging.FromContext(ctx)
	ao.Status.InitializeConditions()
	ao.Status.ObservedGeneration = ao.GetGeneration()

	if !ao.Status.IsNamespaceDelegatesReady() {
		logger.Info("Creating delegates as needed.")
		r.createDelegates(ctx, ao)
	}

	logger.Info("Checking delegates")
	var notReady []string
	for _, delegate := range ao.Status.Delegates {
		d, err := r.operandGetter.ActiveOperands(delegate.Namespace).Get(ctx, ao.GetName(), metav1.GetOptions{})
		if err != nil || !d.Status.IsReady() {
			notReady = append(notReady, fmt.Sprintf("ns: %s, err %+v (or not ready)", delegate.Namespace, err))
		}
	}
	if len(notReady) == 0 {
		logger.Info("All delegates ready")
		ao.Status.MarkNamespaceDelegatesReady()
	} else {
		logger.Info("All delegates ready failed.")
		ao.Status.MarkNamespaceDelegatesReadyFailed(strings.Join(notReady, ", "))
	}

	logger.Info("Injecting ownerrefs")
	err := r.OwnerHandler.HandleOwnerRefs(ctx, kmeta.NewControllerRef(ao), ao.Status.ClusterLive, ao)
	if err != nil {
		logger.Info("Injection failed")
		ao.Status.MarkOwnerRefsInjectedFailed(fmt.Sprintf("%+v", err))
	} else {
		logger.Info("Injection success")
		ao.Status.MarkOwnerRefsInjected()
	}
	return nil
}

func (r clusterReconciler) createDelegates(ctx context.Context, ao *v1alpha1.ClusterActiveOperand) {
	byNamespace := make(map[string][]v1alpha1.LiveRef)
	for _, ref := range ao.Spec.Live {
		temp := &v1alpha1.LiveRef{}
		ref.DeepCopyInto(temp)
		byNamespace[ref.Namespace] = append(byNamespace[ref.Namespace], *temp)
	}
	nss := make([]string, 0, len(byNamespace))
	for ns := range byNamespace {
		nss = append(nss, ns)
	}
	// Consistent ordering for testing.
	sort.Strings(nss)
	delegates := []v1alpha1.DelegateRef{}
	for _, ns := range nss {
		nsLive := byNamespace[ns]
		if ns == "" {
			ao.Status.SetClusterLive(nsLive...)
			continue
		}
		// Make our delegates.
		delegates = append(delegates, v1alpha1.DelegateRef{Namespace: ns})
		_, err := r.operandGetter.ActiveOperands(ns).Create(ctx, &v1alpha1.ActiveOperand{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      ao.GetName(),
			},
			Spec: v1alpha1.ActiveOperandSpec{
				Live: nsLive,
			},
		}, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			ao.Status.MarkNamespaceDelegatesReadyFailed(fmt.Sprintf("Ns %s failed with error %+v", ns, err))
		}
	}
	ao.Status.SetDelegates(delegates...)
}

func (r clusterReconciler) FinalizeKind(ctx context.Context, ao *v1alpha1.ClusterActiveOperand) pkgreconciler.Event {
	logger := logging.FromContext(ctx)
	ao.Status.InitializeConditions()
	ao.Status.ObservedGeneration = ao.Generation

	logger.Info("Deleting delegates")
	if len(ao.Status.Delegates) == 0 {
		return nil
	}
	notDeleted := []string{}
	for _, delegate := range ao.Status.Delegates {
		if err := r.operandGetter.ActiveOperands(delegate.Namespace).Delete(ctx, ao.GetName(), deleteOptions()); err != nil && !apierrors.IsNotFound(err) {
			notDeleted = append(notDeleted, fmt.Sprintf("ns: %s, error %+v", delegate.Namespace, err))
		}
	}
	if len(notDeleted) > 0 {
		return fmt.Errorf("Failed to delete delegates [%s]", strings.Join(notDeleted, " ,"))
	}
	return nil
}
