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
	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	apitesting "github.com/knative/pkg/apis/testing"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func TestSpaceDuckTypes(t *testing.T) {
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
	space := Space{}
	testutil.AssertEqual(t, "empty space generation", int64(0), space.GetGeneration())

	answer := int64(42)
	space.SetGeneration(answer)
	testutil.AssertEqual(t, "GetGeneration", answer, space.GetGeneration())
}

func TestSpaceIsReady(t *testing.T) {
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

	return status
}

func TestSpaceHappyPath(t *testing.T) {
	status := initTestStatus(t)

	status.PropagateDeveloperRoleStatus(nil)
	status.PropagateAuditorRoleStatus(nil)
	status.PropagateNamespaceStatus(&v1.Namespace{Status: v1.NamespaceStatus{Phase: v1.NamespaceActive}})

	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionNamespaceReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionAuditorRoleReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), SpaceConditionDeveloperRoleReady, t)
}

func TestPropagateNamespaceStatus_terminating(t *testing.T) {
	status := initTestStatus(t)

	status.PropagateNamespaceStatus(&v1.Namespace{Status: v1.NamespaceStatus{Phase: v1.NamespaceTerminating}})

	apitesting.CheckConditionFailed(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionFailed(status.duck(), SpaceConditionNamespaceReady, t)
}

func TestSpaceStatus_lifecycle(t *testing.T) {
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
				status.PropagateNamespaceStatus(&v1.Namespace{Status: v1.NamespaceStatus{Phase: v1.NamespaceActive}})
			},
			ExpectSucceeded: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
				SpaceConditionAuditorRoleReady,
				SpaceConditionDeveloperRoleReady,
			},
		},
		"terminating namespace": {
			Init: func(status *SpaceStatus) {
				status.PropagateNamespaceStatus(&v1.Namespace{Status: v1.NamespaceStatus{Phase: v1.NamespaceTerminating}})
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
				status.PropagateNamespaceStatus(&v1.Namespace{Status: v1.NamespaceStatus{}})
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
