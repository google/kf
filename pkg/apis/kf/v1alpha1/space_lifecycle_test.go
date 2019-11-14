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

package v1alpha1

import (
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
)

func TestSpaceDuckTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    duck.Implementable
	}{
		{
			name: "conditions",
			t:    &duckv1beta1.Conditions{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := duck.VerifyType(&Space{}, test.t)
			if err != nil {
				t.Errorf("VerifyType(Service, %T) = %v", test.t, err)
			}
		})
	}
}

func TestSpaceGeneration(t *testing.T) {
	t.Parallel()
	space := Space{}
	testutil.AssertEqual(t, "empty space generation", int64(0), space.GetGeneration())

	answer := int64(42)
	space.SetGeneration(answer)
	testutil.AssertEqual(t, "GetGeneration", answer, space.GetGeneration())
}

func TestSpaceIsReady(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		status  SpaceStatus
		isReady bool
	}{{
		name:    "empty status should not be ready",
		status:  SpaceStatus{},
		isReady: false,
	}, {
		name: "Different condition type should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: false,
	}, {
		name: "False condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Unknown condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Missing condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type: SpaceConditionReady,
				}},
			},
		},
		isReady: false,
	}, {
		name: "True condition status should be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status should be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SpaceConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status false should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SpaceConditionReady,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isReady: false,
	}}

	for _, tc := range cases {
		testutil.AssertEqual(t, tc.name, tc.isReady, tc.status.IsReady())
	}
}

func initTestStatus(t *testing.T) *SpaceStatus {
	t.Helper()
	status := &SpaceStatus{}
	status.InitializeConditions()

	// sanity check
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionNamespaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionAuditorRoleReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionDeveloperRoleReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionResourceQuotaReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionLimitRangeReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionBuildServiceAccountReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionBuildSecretReady, t)

	return status
}

func TestSpaceHappyPath(t *testing.T) {
	t.Parallel()
	status := initTestStatus(t)
	status.PropagateDeveloperRoleStatus(nil)
	status.PropagateAuditorRoleStatus(nil)
	status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}})
	status.PropagateResourceQuotaStatus(&corev1.ResourceQuota{
		Status: corev1.ResourceQuotaStatus{},
	})
	status.PropagateLimitRangeStatus(nil)
	status.PropagateBuildServiceAccountStatus(nil)
	status.PropagateBuildSecretStatus(nil)

	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionNamespaceReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionAuditorRoleReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionDeveloperRoleReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionResourceQuotaReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionLimitRangeReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionBuildServiceAccountReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionBuildSecretReady, t)
}

func TestPropagateNamespaceStatus_terminating(t *testing.T) {
	t.Parallel()
	status := initTestStatus(t)

	status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating}})

	apitesting.CheckConditionFailed(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionFailed(status.duck(), SpaceConditionNamespaceReady, t)
}

func TestPropagateResourceQuotaStatus(t *testing.T) {
	t.Parallel()
	status := initTestStatus(t)

	memHard, _ := resource.ParseQuantity("20Gi")
	cpuHard, _ := resource.ParseQuantity("800m")
	memUsed, _ := resource.ParseQuantity("1Gi")
	cpuUsed, _ := resource.ParseQuantity("100m")
	hard := corev1.ResourceList{
		corev1.ResourceMemory: memHard,
		corev1.ResourceCPU:    cpuHard,
	}
	used := corev1.ResourceList{
		corev1.ResourceMemory: memUsed,
		corev1.ResourceCPU:    cpuUsed,
	}
	quotaToPropagate := &corev1.ResourceQuota{
		Status: corev1.ResourceQuotaStatus{
			Hard: hard,
			Used: used,
		},
	}
	status.PropagateResourceQuotaStatus(quotaToPropagate)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionResourceQuotaReady, t)
	testutil.AssertEqual(t, "quota status", quotaToPropagate.Status, status.Quota)
}

func TestSpaceStatus_lifecycle(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		Init func(*SpaceStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *SpaceStatus) {
				status.PropagateDeveloperRoleStatus(nil)
				status.PropagateAuditorRoleStatus(nil)
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}})
				status.PropagateResourceQuotaStatus(&corev1.ResourceQuota{
					Status: corev1.ResourceQuotaStatus{},
				})
				status.PropagateLimitRangeStatus(nil)
				status.PropagateBuildServiceAccountStatus(nil)
				status.PropagateBuildSecretStatus(nil)
			},
			ExpectSucceeded: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
				SpaceConditionAuditorRoleReady,
				SpaceConditionDeveloperRoleReady,
				SpaceConditionResourceQuotaReady,
				SpaceConditionLimitRangeReady,
				SpaceConditionBuildServiceAccountReady,
				SpaceConditionBuildSecretReady,
			},
		},
		"terminating namespace": {
			Init: func(status *SpaceStatus) {
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating}})
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionAuditorRoleReady,
				SpaceConditionDeveloperRoleReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
			},
		},
		"unknown namespace": {
			Init: func(status *SpaceStatus) {
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{}})
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionAuditorRoleReady,
				SpaceConditionDeveloperRoleReady,
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
			},
		},
		"ns not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkNamespaceNotOwned("my-ns")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionAuditorRoleReady,
				SpaceConditionDeveloperRoleReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
			},
		},
		"dev role not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkDeveloperRoleNotOwned("my-devrole")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionAuditorRoleReady,
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionDeveloperRoleReady,
			},
		},
		"auditor role not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkAuditorRoleNotOwned("my-auditorrole")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionDeveloperRoleReady,
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionAuditorRoleReady,
			},
		},
		"resource quota not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkResourceQuotaNotOwned("space-quota")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionResourceQuotaReady,
			},
		},
		"limit range not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkLimitRangeNotOwned("space-limit-range")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionLimitRangeReady,
			},
		},
		"Build ServiceAccount not owned": {
			Init: func(status *SpaceStatus) {
				status.MarkBuildServiceAccountNotOwned("build-service-account")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionBuildServiceAccountReady,
			},
		},
		"Build Secret not owned": {
			Init: func(status *SpaceStatus) {
				err := status.BuildSecretCondition().MarkChildNotOwned("build-secret")
				testutil.AssertEqual(t, "error", "There is an existing Build Secret \"build-secret\" that we do not own.", err.Error())
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionBuildSecretReady,
			},
		},
	}

	// XXX: if we start copying state from subresources back to the parent,
	// ensure that the state is updated.

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestStatus(t)

			tc.Init(status)

			for _, exp := range tc.ExpectFailed {
				apitesting.CheckConditionFailed(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectOngoing {
				apitesting.CheckConditionOngoing(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectSucceeded {
				apitesting.CheckConditionSucceeded(status.duck(), exp, t)
			}
		})
	}
}
