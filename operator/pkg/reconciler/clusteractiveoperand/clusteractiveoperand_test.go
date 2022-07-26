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
	"fmt"
	"testing"
	"time"

	"kf-operator/pkg/apis/operand/v1alpha1"
	clientset "kf-operator/pkg/client/injection/client"
	operatorclient "kf-operator/pkg/client/injection/client"
	clusteractiveoperandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/clusteractiveoperand"
	"kf-operator/pkg/operand/injection/ownerhandler"
	fakeownerhandler "kf-operator/pkg/operand/injection/ownerhandler/fake"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	rtesting "knative.dev/pkg/reconciler/testing"

	_ "kf-operator/pkg/client/injection/informers/operand/v1alpha1/clusteractiveoperand/fake"
	_ "kf-operator/pkg/operand/injection/dynamichelper"
	. "kf-operator/pkg/reconciler/testing"
	. "kf-operator/pkg/testing/operand/v1alpha1"

	. "knative.dev/pkg/logging/testing"
)

var (
	clusterref = CreateLiveRef("blah", "", appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind())
)

func nsRef(ns string) v1alpha1.LiveRef {
	return CreateLiveRef("blah", ns, appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind())
}

func TestReconcile(t *testing.T) {
	table := rtesting.TableTest{
		{
			Name:        "empty operand trivially ready",
			Key:         "test-obj",
			Objects:     []runtime.Object{ClusterActiveOperandWithDefaults("test-obj")},
			WantPatches: []clientgotesting.PatchActionImpl{ClusterActiveOperandPatchFinalizer("test-obj")},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: ClusterActiveOperandWithDefaults("test-obj", WithClusterLiveRefs(), ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
				},
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx:                     SuccessOwnerInjectContext(t),
		}, {
			Name:        "just namespace refs, created not ready",
			Key:         "test-obj",
			Objects:     []runtime.Object{ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(nsRef("ns1")))},
			WantPatches: []clientgotesting.PatchActionImpl{ClusterActiveOperandPatchFinalizer("test-obj")},
			WantCreates: []runtime.Object{
				ActiveOperand("test-obj", "ns1", WithLiveRefs(nsRef("ns1"))),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(nsRef("ns1")), WithDelegates("ns1"), WithClusterLiveRefs(), ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReadyFailed("ns: ns1, err <nil> (or not ready)")),
				},
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx:                     SuccessOwnerInjectContext(t),
		}, {
			Name:        "just namespace refs, ready",
			Key:         "test-obj",
			Objects:     []runtime.Object{ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(nsRef("ns1"))), ActiveOperandWithDefaults("test-obj", "ns1", WithLiveRefs(nsRef("ns1")), WithOwnerRefsInjected())},
			WantPatches: []clientgotesting.PatchActionImpl{ClusterActiveOperandPatchFinalizer("test-obj")},
			WantCreates: []runtime.Object{
				ActiveOperand("test-obj", "ns1", WithLiveRefs(nsRef("ns1"))),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(nsRef("ns1")), WithDelegates("ns1"), WithClusterLiveRefs(), ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
				},
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx:                     SuccessOwnerInjectContext(t),
		}, {
			Name:        "just cluster refs, injection failed",
			Key:         "test-obj",
			Objects:     []runtime.Object{ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(clusterref))},
			WantPatches: []clientgotesting.PatchActionImpl{ClusterActiveOperandPatchFinalizer("test-obj")},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{Object: ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(clusterref), WithClusterLiveRefs(clusterref), ClusterWithOwnerRefsInjectedFailed("Failed with broken"), WithNamespaceDelegatesReady())},
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx:                     FailOwnerInjection(t, "broken"),
		}, {
			Name:        "just cluster refs, ready",
			Key:         "test-obj",
			Objects:     []runtime.Object{ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(clusterref))},
			WantPatches: []clientgotesting.PatchActionImpl{ClusterActiveOperandPatchFinalizer("test-obj")},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{Object: ClusterActiveOperandWithDefaults("test-obj", ClusterWithLiveRefs(clusterref), WithClusterLiveRefs(clusterref), ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady())},
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx:                     SuccessOwnerInjectContext(t),
		},
	}

	factory := MakeFactory(func(ctx context.Context, listers *Listers) controller.Reconciler {
		return clusteractiveoperandreconciler.NewReconciler(ctx,
			logging.FromContext(ctx),
			operatorclient.Get(ctx),
			listers.GetClusterActiveOperandLister(),
			controller.GetEventRecorder(ctx),
			&clusterReconciler{
				OwnerHandler:  ownerhandler.Get(ctx),
				operandGetter: clientset.Get(ctx).OperandV1alpha1(),
				enqueueAfter:  func(interface{}, time.Duration) {}})
	})

	table.Test(t, factory)
}

func SuccessOwnerInjectContext(t *testing.T) context.Context {
	ctx, _ := fakeownerhandler.With(TestContextWithLogger(t))
	return ctx
}

func FailOwnerInjection(t *testing.T, msg string) context.Context {
	ctx, fakeOwnerHandler := fakeownerhandler.With(TestContextWithLogger(t))
	fakeOwnerHandler.SetError(fmt.Errorf("Failed with %s", msg))
	return ctx
}
