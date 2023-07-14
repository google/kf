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
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
)

func TestBuildDuckTypes(t *testing.T) {
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
			err := duck.VerifyType(&Build{}, test.t)
			if err != nil {
				t.Errorf("VerifyType(Service, %T) = %v", test.t, err)
			}
		})
	}
}

func TestBuildGeneration(t *testing.T) {
	space := Build{}
	testutil.AssertEqual(t, "empty space generation", int64(0), space.GetGeneration())

	answer := int64(42)
	space.SetGeneration(answer)
	testutil.AssertEqual(t, "GetGeneration", answer, space.GetGeneration())
}

func TestBuildSucceeded(t *testing.T) {
	cases := []struct {
		name        string
		status      BuildStatus
		isSucceeded bool
	}{{
		name:        "empty status should not be succeeded",
		status:      BuildStatus{},
		isSucceeded: false,
	}, {
		name: "Different condition type should not be succeeded",
		status: BuildStatus{
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
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   BuildConditionSucceeded,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "Unknown condition status should not be succeeded",
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   BuildConditionSucceeded,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "Missing condition status should not be succeeded",
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type: BuildConditionSucceeded,
				}},
			},
		},
		isSucceeded: false,
	}, {
		name: "True condition status should be succeeded",
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   BuildConditionSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isSucceeded: true,
	}, {
		name: "Multiple conditions with succeeded status should be succeeded",
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   BuildConditionSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isSucceeded: true,
	}, {
		name: "Multiple conditions with succeeded status false should not be succeeded",
		status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   BuildConditionSucceeded,
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

func initTestBuildStatus(t *testing.T) *BuildStatus {
	t.Helper()
	status := &BuildStatus{}
	status.InitializeConditions()

	// sanity check
	apitesting.CheckConditionOngoing(status.duck(), BuildConditionSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), BuildConditionTaskRunReady, t)
	apitesting.CheckConditionOngoing(status.duck(), BuildConditionSourcePackageReady, t)

	return status
}

func happySourcePackage() *SourcePackage {
	base := &SourcePackage{
		Status: SourcePackageStatus{},
	}
	base.Status.Conditions = duckv1beta1.Conditions{
		{Type: apis.ConditionSucceeded, Status: corev1.ConditionTrue},
	}
	return base
}

func unreconciledTaskRun() *tektonv1beta1.TaskRun {
	return &tektonv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-build-name",
		},
		Spec: tektonv1beta1.TaskRunSpec{},
	}
}

func pendingTaskRun() *tektonv1beta1.TaskRun {
	base := unreconciledTaskRun()

	base.Status.Conditions = duckv1beta1.Conditions{
		{Type: apis.ConditionSucceeded, Status: corev1.ConditionUnknown},
	}

	startTime := metav1.NewTime(time.Unix(0, 0))
	base.Status.StartTime = &startTime

	return base
}

func happyTaskRun() *tektonv1beta1.TaskRun {
	base := pendingTaskRun()

	base.Status.Conditions = duckv1beta1.Conditions{
		{Type: apis.ConditionSucceeded, Status: corev1.ConditionTrue},
	}

	endTime := metav1.NewTime(time.Unix(1000, 0))
	base.Status.CompletionTime = &endTime

	return base
}

func failedTaskRun() *tektonv1beta1.TaskRun {
	base := pendingTaskRun()

	base.Status.Conditions = duckv1beta1.Conditions{
		{Type: apis.ConditionSucceeded, Status: corev1.ConditionFalse},
	}

	endTime := metav1.NewTime(time.Unix(1000, 0))
	base.Status.CompletionTime = &endTime

	return base
}

func happySecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-secret-name",
		},
	}
}

func TestBuildHappyPath(t *testing.T) {
	status := initTestBuildStatus(t)

	// Build starts out pending while the container kicks off
	status.PropagateBuildStatus(pendingTaskRun())

	apitesting.CheckConditionOngoing(status.duck(), BuildConditionSucceeded, t)
	apitesting.CheckConditionOngoing(status.duck(), BuildConditionSpaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), BuildConditionSourcePackageReady, t)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.BuildName)
	testutil.AssertEqual(t, "Image", "", status.Image) // Image not populated until build succeeds

	// Space is healthy.
	status.MarkSpaceHealthy()

	// SourcePackage succeeds.
	status.PropagateSourcePackageStatus(happySourcePackage())

	// Build succeeds.
	status.PropagateBuildStatus(happyTaskRun())

	apitesting.CheckConditionSucceeded(status.duck(), BuildConditionSucceeded, t)
	apitesting.CheckConditionSucceeded(status.duck(), BuildConditionTaskRunReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), BuildConditionSourcePackageReady, t)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.BuildName)
	testutil.AssertEqual(t, "Image", "some-container-image", status.Image)
}

func TestBuildStatus_lifecycle(t *testing.T) {
	cases := map[string]struct {
		Init func(*BuildStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *BuildStatus) {
				status.MarkSpaceHealthy()
				status.PropagateBuildStatus(happyTaskRun())
				status.PropagateSourcePackageStatus(happySourcePackage())
			},
			ExpectSucceeded: []apis.ConditionType{
				BuildConditionSpaceReady,
				BuildConditionSucceeded,
				BuildConditionTaskRunReady,
				BuildConditionSourcePackageReady,
			},
		},
		"task run not owned": {
			Init: func(status *BuildStatus) {
				condition := status.TaskRunCondition()
				condition.MarkChildNotOwned("my-build")
			},
			ExpectOngoing: []apis.ConditionType{},
			ExpectFailed: []apis.ConditionType{
				BuildConditionSucceeded,
				BuildConditionTaskRunReady,
			},
		},
	}

	// XXX: if we start copying state from subresources back to the parent,
	// ensure that the state is updated.

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestBuildStatus(t)

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

func TestBuildStatus_PropagateBuildStatus(t *testing.T) {
	cases := map[string]struct {
		build *tektonv1beta1.TaskRun
	}{
		"nil": {
			build: nil,
		},
		"unreconciled": {
			build: unreconciledTaskRun(),
		},
		"ongoing": {
			build: pendingTaskRun(),
		},
		"completed": {
			build: happyTaskRun(),
		},
		"failed": {
			build: failedTaskRun(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := BuildStatus{}
			status.PropagateBuildStatus(tc.build)

			// Remove nondeterministic times for golden tests
			for i := range status.Conditions {
				status.Conditions[i].LastTransitionTime = apis.VolatileTime{}
			}

			testutil.AssertGoldenJSONContext(t, "buildstatus", status, map[string]interface{}{
				"TaskRun": tc.build,
			})
		})
	}
}
