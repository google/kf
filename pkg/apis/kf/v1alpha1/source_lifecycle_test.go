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
	build "github.com/google/kf/third_party/knative-build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
)

func TestSourceDuckTypes(t *testing.T) {
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
			err := duck.VerifyType(&Source{}, test.t)
			if err != nil {
				t.Errorf("VerifyType(Service, %T) = %v", test.t, err)
			}
		})
	}
}

func TestSourceGeneration(t *testing.T) {
	space := Source{}
	testutil.AssertEqual(t, "empty space generation", int64(0), space.GetGeneration())

	answer := int64(42)
	space.SetGeneration(answer)
	testutil.AssertEqual(t, "GetGeneration", answer, space.GetGeneration())
}

func TestSourceSucceeded(t *testing.T) {
	cases := []struct {
		name        string
		status      SourceStatus
		isSucceeded bool
	}{{
		name:        "empty status should not be succeeded",
		status:      SourceStatus{},
		isSucceeded: false,
	}, {
		name: "Different condition type should not be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "False condition status should not be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SourceConditionSucceeded,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "Unknown condition status should not be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SourceConditionSucceeded,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "Missing condition status should not be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type: SourceConditionSucceeded,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "True condition status should be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SourceConditionSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isSucceeded: true,
	}, {
		name: "Multiple conditions with succeeded status should be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SourceConditionSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isSucceeded: true,
	}, {
		name: "Multiple conditions with succeeded status false should not be succeeded",
		status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SourceConditionSucceeded,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isSucceeded: false,
	}}

	for _, tc := range cases {
		testutil.AssertEqual(t, tc.name, tc.isSucceeded, tc.status.Succeeded())
	}
}

func initTestSourceStatus(t *testing.T) *SourceStatus {
	t.Helper()
	status := &SourceStatus{}
	status.InitializeConditions()

	// sanity check
	apitesting.CheckConditionOngoing(status.duck(), SourceConditionSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), SourceConditionBuildSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), SourceConditionBuildSecretReady, t)

	return status
}

func happyBuild() *build.Build {
	return &build.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-build-name",
		},
		Spec: build.BuildSpec{
			Template: &build.TemplateInstantiationSpec{
				Arguments: []build.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: "some-container-image",
					},
				},
			},
		},
		Status: build.BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   apis.ConditionSucceeded,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}
}

func happySecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-secret-name",
		},
	}
}

func pendingBuild() *build.Build {
	return &build.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-build-name",
		},
		Spec: build.BuildSpec{
			Template: &build.TemplateInstantiationSpec{
				Arguments: []build.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: "some-container-image",
					},
				},
			},
		},
		Status: build.BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   apis.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
		},
	}
}

func TestSourceHappyPath(t *testing.T) {
	status := initTestSourceStatus(t)

	// Build starts out pending while the container kicks off
	status.PropagateBuildStatus(pendingBuild())

	apitesting.CheckConditionOngoing(status.duck(), SourceConditionSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), SourceConditionBuildSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), SourceConditionBuildSecretReady, t)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.BuildName)
	testutil.AssertEqual(t, "Image", "", status.Image) // Image not populated until build succeeds

	// Build succeeds
	status.PropagateBuildStatus(happyBuild())
	status.PropagateBuildSecretStatus(happySecret())

	apitesting.CheckConditionSucceeded(status.duck(), SourceConditionSucceeded, t)
	apitesting.CheckConditionSucceeded(status.duck(), SourceConditionSucceeded, t)
	apitesting.CheckConditionSucceeded(status.duck(), SourceConditionBuildSucceeded, t)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.BuildName)
	testutil.AssertEqual(t, "Image", "some-container-image", status.Image)
}

func TestSourceStatus_lifecycle(t *testing.T) {
	cases := map[string]struct {
		Init func(*SourceStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *SourceStatus) {
				status.PropagateBuildStatus(happyBuild())
				status.PropagateBuildSecretStatus(happySecret())
			},
			ExpectSucceeded: []apis.ConditionType{
				SourceConditionSucceeded,
				SourceConditionBuildSucceeded,
				SourceConditionBuildSecretReady,
			},
		},
		"happy path with nil secret": {
			Init: func(status *SourceStatus) {
				status.PropagateBuildStatus(happyBuild())
				status.PropagateBuildSecretStatus(nil)
			},
			ExpectSucceeded: []apis.ConditionType{
				SourceConditionSucceeded,
				SourceConditionBuildSucceeded,
				SourceConditionBuildSecretReady,
			},
		},
		"build not owned": {
			Init: func(status *SourceStatus) {
				condition := status.BuildCondition()
				condition.MarkChildNotOwned("my-build")
			},
			ExpectOngoing: []apis.ConditionType{},
			ExpectFailed: []apis.ConditionType{
				SourceConditionSucceeded,
				SourceConditionBuildSucceeded,
			},
		},
		"secret not owned": {
			Init: func(status *SourceStatus) {
				condition := status.BuildSecretCondition()
				condition.MarkChildNotOwned("my-secret")
			},
			ExpectOngoing: []apis.ConditionType{},
			ExpectFailed: []apis.ConditionType{
				SourceConditionSucceeded,
				SourceConditionBuildSecretReady,
			},
		},
	}

	// XXX: if we start copying state from subresources back to the parent,
	// ensure that the state is updated.

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestSourceStatus(t)

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
