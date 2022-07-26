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
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	v1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"
	operatorclient "kf-operator/pkg/client/injection/client"
	operandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/operand"
	poperand "kf-operator/pkg/operand"
	"kf-operator/pkg/operand/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	rtesting "knative.dev/pkg/reconciler/testing"

	_ "kf-operator/pkg/client/injection/client/fake"
	_ "kf-operator/pkg/client/injection/informers/operand/v1alpha1/operand/fake"
	_ "kf-operator/pkg/operand/injection/dynamichelper/fake"
	. "kf-operator/pkg/reconciler/testing"

	. "kf-operator/pkg/testing/k8s"
	. "kf-operator/pkg/testing/operand/v1alpha1"

	. "knative.dev/pkg/logging/testing"
)

var (
	ref = kmeta.NewControllerRef(OperandWithDefaults("test-obj"))
)

type expectation interface {
	expect(*mock.MockResourceReconciler) *gomock.Call
}

type expectApply struct {
	resources []unstructured.Unstructured
	err       error
}

func (e expectApply) expect(m *mock.MockResourceReconciler) *gomock.Call {
	return m.EXPECT().Apply(gomock.Any(), matchResources(e.resources)).Return(e.err)
}

type expectGetState struct {
	resources []unstructured.Unstructured
	result    string
	err       error
}

func (e expectGetState) expect(m *mock.MockResourceReconciler) *gomock.Call {
	return m.EXPECT().GetState(gomock.Any(), matchResources(e.resources)).Return(e.result, e.err)
}

func cmpResources(x, y unstructured.Unstructured) bool {
	if x.GetNamespace() != y.GetNamespace() {
		return x.GetNamespace() < y.GetNamespace()
	}
	if x.GetName() != y.GetName() {
		return x.GetName() < y.GetName()
	}
	return x.GroupVersionKind().String() < y.GroupVersionKind().String()
}

type matchResources []unstructured.Unstructured

func (m matchResources) Matches(x interface{}) bool {
	return cmp.Equal(([]unstructured.Unstructured)(m), x, cmpopts.EquateEmpty(), cmpopts.SortSlices(cmpResources))
}

func (m matchResources) String() string {
	return fmt.Sprintf("%v", ([]unstructured.Unstructured)(m))
}

func (m matchResources) Got(x interface{}) string {
	diff := cmp.Diff(([]unstructured.Unstructured)(m), x, cmpopts.EquateEmpty(), cmpopts.SortSlices(cmpResources))
	if diff != "" {
		return fmt.Sprintf("Diff (-want +got): %s", diff)
	}
	return fmt.Sprintf("%v", x)
}

func TestReconcile(t *testing.T) {
	table := rtesting.TableTest{
		{
			Name:    "no cao, cao created",
			Key:     "test-obj",
			Objects: []runtime.Object{OperandWithDefaults("test-obj", Generations(1, 0))},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithResetLatestCreatedActiveOperand("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			WantCreates: []runtime.Object{
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c"),
			},
			SkipNamespaceValidation: true,
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			Ctx: TestContextWithLogger(t),
		},
		{
			Name: "operand with ready cao",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj", Generations(1, 0)),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithLatestActiveOperandReady("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand, no cao ready",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj", Generations(1, 0), WithSteadyState(t, Deployment("platoon"))),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon")),
						WithResetLatestCreatedActiveOperand("sha256-5989f51667"),
						WithOperandInstallNotReady("PENDING_CHANGES"),
					),
				},
			},
			WantCreates: []runtime.Object{
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-5989f51667", ClusterWithLiveRefs(CreateLiveRef("platoon", "test", appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind()))),
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{resources: ToUnstructured(Deployment("platoon"))},
					expectGetState{resources: ToUnstructured(Deployment("platoon")), result: poperand.PendingChanges},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand post-install not ready",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 0),
					WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
					WithPostInstall(t, Deployment("legion")),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-98458717a8", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
						WithPostInstall(t, Deployment("legion")),
						WithOperandPostInstallNotReady("PENDING_CHANGES"),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"))},
					expectGetState{resources: ToUnstructured(Deployment("platoon"), Deployment("militia")), result: poperand.Installed},
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion"))},
					expectGetState{
						resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion")),
						result:    poperand.PendingChanges,
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand post-install failed",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 0),
					WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
					WithPostInstall(t, Deployment("legion")),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-98458717a8", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
						WithPostInstall(t, Deployment("legion")),
						WithOperandPostInstallFailed(errors.New("PostInstallError")),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			WantEvents: []string{rtesting.Eventf(corev1.EventTypeWarning, "InternalError", "PostInstallError")},
			WantErr:    true,
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"))},
					expectGetState{resources: ToUnstructured(Deployment("platoon"), Deployment("militia")), result: poperand.Installed},
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion"))},
					expectGetState{
						resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion")),
						result:    "", err: errors.New("PostInstallError"),
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand ready",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 0),
					WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-f9ceba1976", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady())},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
						WithLatestActiveOperandReady("sha256-f9ceba1976"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"))},
					expectGetState{resources: ToUnstructured(Deployment("platoon"), Deployment("militia")), result: poperand.Installed},
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"))},
					expectGetState{
						resources: ToUnstructured(Deployment("platoon"), Deployment("militia")),
						result:    poperand.Installed,
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand, same generation, skips installed SteadyState",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 1),
					WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
					WithPostInstall(t, Deployment("legion")),
					WithResetLatestCreatedActiveOperand("sha256-98458717a8"),
					WithInstalledSteadyStateGeneration(1),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-98458717a8", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
						WithPostInstall(t, Deployment("legion")),
						WithLatestActiveOperandCreated("sha256-98458717a8"),
						WithLatestActiveOperandReady("sha256-98458717a8"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion"))},
					expectGetState{
						resources: ToUnstructured(Deployment("platoon"), Deployment("militia"), Deployment("legion")),
						result:    poperand.Installed,
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand merges post-install resources",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 1),
					WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
					WithPostInstall(t,
						Deployment("platoon",
							WithDeploymentContainer(&corev1.Container{
								Name: "new-container",
							}),
						),
						Deployment("legion")),
					WithResetLatestCreatedActiveOperand("sha256-98458717a8"),
					WithInstalledSteadyStateGeneration(1),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-98458717a8", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t, Deployment("platoon"), Deployment("militia")),
						WithPostInstall(t,
							Deployment("platoon",
								WithDeploymentContainer(&corev1.Container{
									Name: "new-container",
								}),
							),
							Deployment("legion"),
						),
						WithLatestActiveOperandCreated("sha256-98458717a8"),
						WithLatestActiveOperandReady("sha256-98458717a8"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{
						resources: ToUnstructured(
							Deployment("platoon",
								WithDeploymentContainer(&corev1.Container{
									Name: "new-container",
								}),
							),
							Deployment("militia"),
							Deployment("legion"),
						)},
					expectGetState{
						resources: ToUnstructured(
							Deployment("platoon",
								WithDeploymentContainer(&corev1.Container{
									Name: "new-container",
								}),
							),
							Deployment("militia"),
							Deployment("legion"),
						),
						result: poperand.Installed,
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "non-empty operand merges post-install resources with new apiversion",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 1),
					WithSteadyState(t,
						&appsv1beta2.Deployment{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "platoon",
								Namespace: "test",
							},
						},
						Deployment("militia"),
					),
					WithPostInstall(t, Deployment("platoon"), Deployment("legion")),
					WithResetLatestCreatedActiveOperand("sha256-98458717a8"),
					WithInstalledSteadyStateGeneration(1),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-98458717a8", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithSteadyState(t,
							&appsv1beta2.Deployment{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "platoon",
									Namespace: "test",
								},
							},
							Deployment("militia"),
						),
						WithPostInstall(t, Deployment("platoon"), Deployment("legion")),
						WithLatestActiveOperandCreated("sha256-98458717a8"),
						WithLatestActiveOperandReady("sha256-98458717a8"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{
						resources: ToUnstructured(
							Deployment("platoon"),
							Deployment("militia"),
							Deployment("legion"),
						)},
					expectGetState{
						resources: ToUnstructured(
							Deployment("platoon"),
							Deployment("militia"),
							Deployment("legion"),
						),
						result: poperand.Installed,
					},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "operand changed, new CAO created, not yet ready",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 0),
					WithLatestActiveOperandCreated("sha256-f9ceba1976"),
					WithLatestActiveOperandReady("sha256-f9ceba1976"),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-f9ceba1976", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithLatestActiveOperandReady("sha256-f9ceba1976"),
						WithResetLatestCreatedActiveOperand("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			WantCreates: []runtime.Object{
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c"),
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name: "operand changed, new CAO created and ready, old ones cleaned up",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults("test-obj",
					Generations(1, 0),
				),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-f9ceba1976", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-f9ceba1977", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithLatestActiveOperandReady("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			WantDeletes: []clientgotesting.DeleteActionImpl{
				{
					ActionImpl: clientgotesting.ActionImpl{
						Namespace: "test",
						Verb:      "delete",
						Resource:  v1alpha1.SchemaGroupVersion.WithResource("clusteractiveoperands"),
					},
					Name: "sha256-f9ceba1976",
				},
				{
					ActionImpl: clientgotesting.ActionImpl{
						Namespace: "test",
						Verb:      "delete",
						Resource:  v1alpha1.SchemaGroupVersion.WithResource("clusteractiveoperands"),
					},
					Name: "sha256-f9ceba1977",
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		{
			Name:    "operand changed, new CAO created, old doesn't exist",
			Key:     "test-obj",
			Objects: []runtime.Object{OperandWithDefaults("test-obj", Generations(1, 0), WithLatestActiveOperandCreated("sha256-f9ceba1976"))},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithResetLatestCreatedActiveOperand("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			WantCreates: []runtime.Object{
				ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c"),
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
		// TODO: remove this test after http://b/201351749
		{
			Name: "operand ready with ready CAO, no owner label, label injected",
			Key:  "test-obj",
			Objects: []runtime.Object{
				OperandWithDefaults(
					"test-obj",
					Generations(1, 1),
					WithLatestActiveOperandCreated("sha256-4f53cda18c"),
					WithLatestActiveOperandReady("sha256-4f53cda18c"),
					WithInstalledSteadyStateGeneration(1),
					WithOperandInstallSuccessful(),
				),
				ClusterActiveOperand(*ref, "sha256-4f53cda18c", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: ClusterActiveOperandWithOwnerLabel(*ref, "sha256-4f53cda18c", ClusterWithOwnerRefsInjected(), WithNamespaceDelegatesReady()),
				},
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: OperandWithDefaults("test-obj",
						Generations(1, 1),
						WithLatestActiveOperandCreated("sha256-4f53cda18c"),
						WithLatestActiveOperandReady("sha256-4f53cda18c"),
						WithOperandInstallSuccessful(),
						WithInstalledSteadyStateGeneration(1),
					),
				},
			},
			OtherTestData: map[string]interface{}{
				"expectations": []expectation{
					expectApply{},
					expectGetState{result: poperand.Installed},
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     TestContextWithLogger(t),
		},
	}

	factory := func(t *testing.T, r *rtesting.TableRow) (controller.Reconciler, rtesting.ActionRecorderList, rtesting.EventList) {
		factoryFunc := MakeFactory(func(ctx context.Context, listers *Listers) controller.Reconciler {
			mockCtrl := gomock.NewController(t)
			resourceR := mock.NewMockResourceReconciler(mockCtrl)
			if exs := GetOtherData(ctx, "expectations"); exs != nil {
				for _, ex := range exs.([]expectation) {
					ex.expect(resourceR)
				}
			} else {
				calls := []*gomock.Call{
					resourceR.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(GetOtherData(ctx, "apply")),
					resourceR.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(GetOtherData(ctx, "get_state"), GetOtherData(ctx, "get_state_err")),
				}
				if st := GetOtherData(ctx, "post_install_get_state"); st != nil {
					calls = append(calls,
						resourceR.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(GetOtherData(ctx, "post_install_apply")),
						resourceR.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(st, GetOtherData(ctx, "post_install_get_state_err")),
					)
				}
				gomock.InOrder(calls...)
			}
			return operandreconciler.NewReconciler(ctx,
				logging.FromContext(ctx),
				operatorclient.Get(ctx),
				listers.GetOperandLister(),
				controller.GetEventRecorder(ctx),
				&reconciler{
					lock:                       &sync.Mutex{},
					resourceReconciler:         resourceR,
					operandClient:              operatorclient.Get(ctx).OperandV1alpha1(),
					clusterActiveOperandLister: listers.GetClusterActiveOperandLister(),
					enqueueAfter:               func(interface{}, time.Duration) {},
				},
				controller.Options{SkipStatusUpdates: true},
			)
		})
		return factoryFunc(t, r)
	}

	table.Test(t, factory)
}
