// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
)

func TestAppSucceeded(t *testing.T) {
	cases := []struct {
		name    string
		status  AppStatus
		isReady bool
	}{{
		name:    "empty status should not be ready",
		status:  AppStatus{},
		isReady: false,
	}, {
		name: "Different condition type should not be ready",
		status: AppStatus{
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
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   AppConditionReady,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Unknown condition status should not be ready",
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   AppConditionReady,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Missing condition status should not be ready",
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type: AppConditionReady,
				}},
			},
		},
		isReady: false,
	}, {
		name: "True condition status should be ready",
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   AppConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status should be ready",
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   AppConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status false should not be ready",
		status: AppStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   AppConditionReady,
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

func initTestAppStatus(t *testing.T) *AppStatus {
	t.Helper()
	status := &AppStatus{}
	status.InitializeConditions()

	// sanity check
	apitesting.CheckConditionOngoing(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionSpaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionSourceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionEnvVarSecretReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionKnativeServiceReady, t)

	return status
}

func happySource() *Source {
	return &Source{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-source-name",
		},
		Spec: SourceSpec{
			ServiceAccount: "builder-account",
			BuildpackBuild: SourceSpecBuildpackBuild{
				Source:           "gcr.io/my-registry/src-mysource",
				Stack:            "cflinuxfs3",
				BuildpackBuilder: "gcr.io/my-registry/my-builder:latest",
				Image:            "gcr.io/my-registry/output:123",
			},
		},
		Status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   SourceConditionSucceeded,
						Status: corev1.ConditionTrue,
					},
				},
			},
			SourceStatusFields: SourceStatusFields{
				BuildName: "some-build-name",
				Image:     "some-container-image",
			},
		},
	}
}

func pendingSource() *Source {
	return &Source{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-source-name",
		},
		Spec: SourceSpec{
			ServiceAccount: "builder-account",
			BuildpackBuild: SourceSpecBuildpackBuild{
				Source:           "gcr.io/my-registry/src-mysource",
				Stack:            "cflinuxfs3",
				BuildpackBuilder: "gcr.io/my-registry/my-builder:latest",
				Image:            "gcr.io/my-registry/output:123",
			},
		},
		Status: SourceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   SourceConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			SourceStatusFields: SourceStatusFields{
				BuildName: "",
				Image:     "",
			},
		},
	}
}

func envVarSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-secret-name",
		},
		Data: map[string][]byte{
			"some-env-name": []byte("some-env-value"),
		},
	}
}

func happyKnativeService() *serving.Service {
	return &serving.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-service-name",
		},
		Status: serving.ServiceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   serving.ServiceConditionReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
			ConfigurationStatusFields: serving.ConfigurationStatusFields{
				LatestReadyRevisionName:   "some-ready-revision-name",
				LatestCreatedRevisionName: "some-created-revision-name",
			},
			RouteStatusFields: serving.RouteStatusFields{
				URL: &apis.URL{
					Host: "example.com",
				},
			},
		},
	}
}

func pendingKnativeService() *serving.Service {
	return &serving.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-service-name",
		},
		Status: serving.ServiceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   serving.ServiceConditionReady,
						Status: corev1.ConditionUnknown,
					},
				},
			},
		},
	}
}

func TestAppHappyPath(t *testing.T) {
	status := initTestAppStatus(t)

	apitesting.CheckConditionOngoing(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionSpaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionSourceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionEnvVarSecretReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionKnativeServiceReady, t)

	// space is healthy
	status.MarkSpaceHealthy()

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionSpaceReady, t)

	// Source starts out pending
	status.PropagateSourceStatus(pendingSource())

	apitesting.CheckConditionOngoing(status.duck(), AppConditionSourceReady, t)
	testutil.AssertEqual(t, "LatestCreatedSourceName", "some-source-name", status.LatestCreatedSourceName)
	testutil.AssertEqual(t, "LatestReadySourceName", "", status.LatestReadySourceName)
	testutil.AssertEqual(t, "BuildName", "", status.SourceStatusFields.BuildName)
	testutil.AssertEqual(t, "Image", "", status.SourceStatusFields.Image)

	// Source succeeds
	status.PropagateSourceStatus(happySource())

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionSourceReady, t)
	testutil.AssertEqual(t, "LatestReadySourceName", "some-source-name", status.LatestReadySourceName)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.SourceStatusFields.BuildName)
	testutil.AssertEqual(t, "Image", "some-container-image", status.SourceStatusFields.Image)

	// envVarSecret exists
	status.PropagateEnvVarSecretStatus(envVarSecret())

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionEnvVarSecretReady, t)

	// Knative Serving starts out pending
	status.PropagateKnativeServiceStatus(pendingKnativeService())

	apitesting.CheckConditionOngoing(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionKnativeServiceReady, t)

	testutil.AssertEqual(t, "LatestReadyRevisionName", "", status.LatestReadyRevisionName)
	testutil.AssertEqual(t, "LatestCreatedRevisionName", "", status.LatestCreatedRevisionName)
	testutil.AssertEqual(t, "RouteStatusFields", serving.RouteStatusFields{}, status.RouteStatusFields)

	// Knative Serving is ready
	status.PropagateKnativeServiceStatus(happyKnativeService())

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionKnativeServiceReady, t)
	testutil.AssertEqual(t, "LatestReadyRevisionName", "some-ready-revision-name", status.LatestReadyRevisionName)
	testutil.AssertEqual(t, "LatestCreatedRevisionName", "some-created-revision-name", status.LatestCreatedRevisionName)
	testutil.AssertEqual(t, "RouteHost", "example.com", status.RouteStatusFields.URL.Host)
}

func TestAppStatus_lifecycle(t *testing.T) {
	cases := map[string]struct {
		Init func(status *AppStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *AppStatus) {
				status.MarkSpaceHealthy()
				status.PropagateSourceStatus(happySource())
				status.PropagateEnvVarSecretStatus(envVarSecret())
				status.PropagateKnativeServiceStatus(happyKnativeService())
			},
			ExpectSucceeded: []apis.ConditionType{
				AppConditionReady,
				AppConditionSpaceReady,
				AppConditionSourceReady,
				AppConditionEnvVarSecretReady,
				AppConditionKnativeServiceReady,
			},
		},
		"stopped app": {
			Init: func(status *AppStatus) {
				status.MarkSpaceHealthy()
				status.PropagateSourceStatus(happySource())
				status.PropagateEnvVarSecretStatus(envVarSecret())
				// Nil Knative service because a stopped app doesn't have a
				// Knative service.
				status.PropagateKnativeServiceStatus(nil)
			},
			ExpectSucceeded: []apis.ConditionType{
				AppConditionSpaceReady,
				AppConditionSourceReady,
				AppConditionEnvVarSecretReady,
			},
		},
		"space unhealthy": {
			Init: func(status *AppStatus) {
				status.MarkSpaceUnhealthy("Terminating", "Namespace is terminating")
			},
			ExpectFailed: []apis.ConditionType{
				AppConditionReady,
				AppConditionSpaceReady,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestAppStatus(t)

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

func TestServiceBindingConditionType(t *testing.T) {
	cases := map[string]struct {
		binding     *servicecatalogv1beta1.ServiceBinding
		expected    apis.ConditionType
		expectedErr error
	}{
		"nil": {
			expectedErr: errors.New("binding cannot be nil"),
		},
		"missing label": {
			binding: &servicecatalogv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-binding",
					Labels: map[string]string{
						"wow": "cool",
					},
				},
			},
			expectedErr: errors.New("binding my-binding is missing the label app.kubernetes.io/component"),
		},
		"correct": {
			binding: &servicecatalogv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-binding",
					Labels: map[string]string{
						"app.kubernetes.io/component": "my-service-instance",
					},
				},
			},
			expected: apis.ConditionType("Ready-my-service-instance"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := ServiceBindingConditionType(tc.binding)
			testutil.AssertEqual(t, "err", tc.expectedErr, err)
			testutil.AssertEqual(t, "conditionType", tc.expected, actual)
		})
	}
}
