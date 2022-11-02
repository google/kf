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

	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"github.com/google/kf/v2/pkg/kf/testutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
	"knative.dev/pkg/ptr"
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

	// sanity check conditions get initiailized as unknown
	for _, cond := range status.Conditions {
		apitesting.CheckConditionOngoing(status.duck(), cond.Type, t)
	}

	// sanity check total conditions (add 1 for "Ready")
	testutil.AssertEqual(t, "conditions count", 10, len(status.Conditions))

	return status
}

func happyBuild() *Build {
	return &Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-build-name",
		},
		Spec: BuildSpec{
			BuildTaskRef: buildpackV3BuildTaskRef(),
		},
		Status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   BuildConditionSucceeded,
						Status: corev1.ConditionTrue,
					},
				},
			},
			BuildStatusFields: BuildStatusFields{
				BuildName: "some-build-name",
				Image:     "some-container-image",
			},
		},
	}
}

func pendingBuild() *Build {
	return &Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-build-name",
		},
		Spec: BuildSpec{
			BuildTaskRef: buildpackV3BuildTaskRef(),
		},
		Status: BuildStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   BuildConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			BuildStatusFields: BuildStatusFields{
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

func serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sa-app-name",
		},
	}
}

func happyDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-service-name",
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 3,
			ReadyReplicas:     3,
			UpdatedReplicas:   3,
			Replicas:          3,
		},
	}
}

func pendingDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-service-name",
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
			ReadyReplicas:     3,
			UpdatedReplicas:   3,
			Replicas:          3,
		},
	}
}

func happyHorizontalPodAutoscaler() *autoscalingv1.HorizontalPodAutoscaler {
	return &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-hpa-name",
		},
		Status: autoscalingv1.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 1,
			DesiredReplicas: 1,
		},
	}
}

func pendingHorizontalPodAutoscaler() *autoscalingv1.HorizontalPodAutoscaler {
	return &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-hpa-name",
		},
		Status: autoscalingv1.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 1,
			DesiredReplicas: 2,
		},
	}
}

func happyService() *corev1.Service {
	return &corev1.Service{}
}

func TestAppHappyPath(t *testing.T) {
	status := initTestAppStatus(t)

	apitesting.CheckConditionOngoing(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionSpaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionBuildReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionEnvVarSecretReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionDeploymentReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionServiceAccountReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionServiceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionHorizontalPodAutoscalerReady, t)

	// space is healthy
	status.MarkSpaceHealthy()

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionSpaceReady, t)

	// bindings become ready
	status.PropagateServiceInstanceBindingsStatus(nil)
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionServiceInstanceBindingsReady, t)

	// Build starts out pending
	status.PropagateBuildStatus(pendingBuild())

	apitesting.CheckConditionOngoing(status.duck(), AppConditionBuildReady, t)
	testutil.AssertEqual(t, "LatestCreatedBuildName", "some-build-name", status.LatestCreatedBuildName)
	testutil.AssertEqual(t, "LatestReadyBuildName", "", status.LatestReadyBuildName)
	testutil.AssertEqual(t, "BuildName", "", status.BuildStatusFields.BuildName)
	testutil.AssertEqual(t, "Image", "", status.BuildStatusFields.Image)

	// Build succeeds
	status.PropagateBuildStatus(happyBuild())

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionBuildReady, t)
	testutil.AssertEqual(t, "LatestReadyBuildName", "some-build-name", status.LatestReadyBuildName)
	testutil.AssertEqual(t, "BuildName", "some-build-name", status.BuildStatusFields.BuildName)
	testutil.AssertEqual(t, "Image", "some-container-image", status.BuildStatusFields.Image)

	// envVarSecret exists
	status.PropagateEnvVarSecretStatus(envVarSecret())
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionEnvVarSecretReady, t)

	// service gets reconciled
	status.PropagateServiceStatus(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "my-namespace",
		},
	})

	// service account gets reconciled
	status.PropagateServiceAccountStatus(serviceAccount())
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionServiceAccountReady, t)

	// Hpa gets reconciled
	status.PropagateAutoscalerV1Status(pendingHorizontalPodAutoscaler())
	apitesting.CheckConditionOngoing(status.duck(), AppConditionHorizontalPodAutoscalerReady, t)

	// Deployment starts out pending
	status.PropagateDeploymentStatus(pendingDeployment())

	apitesting.CheckConditionOngoing(status.duck(), AppConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), AppConditionDeploymentReady, t)

	// Deployment is ready
	status.PropagateDeploymentStatus(happyDeployment())
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionDeploymentReady, t)

	// Autoscaler is ready
	status.PropagateAutoscalerV1Status(happyHorizontalPodAutoscaler())
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionHorizontalPodAutoscalerReady, t)

	// Routes and bindings are reeady
	status.PropagateRouteStatus(nil, nil, nil)
	apitesting.CheckConditionSucceeded(status.duck(), AppConditionRouteReady, t)

	apitesting.CheckConditionSucceeded(status.duck(), AppConditionReady, t)
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
				status.PropagateBuildStatus(happyBuild())
				status.PropagateEnvVarSecretStatus(envVarSecret())
				status.PropagateServiceStatus(happyService())
				status.PropagateRouteStatus(nil, nil, nil)
				status.PropagateServiceInstanceBindingsStatus(nil)
				status.PropagateServiceAccountStatus(serviceAccount())
				status.PropagateDeploymentStatus(happyDeployment())
				status.PropagateAutoscalerV1Status(happyHorizontalPodAutoscaler())
			},
			ExpectSucceeded: []apis.ConditionType{
				AppConditionReady,
				AppConditionSpaceReady,
				AppConditionBuildReady,
				AppConditionEnvVarSecretReady,
				AppConditionServiceReady,
				AppConditionRouteReady,
				AppConditionServiceInstanceBindingsReady,
				AppConditionServiceAccountReady,
				AppConditionDeploymentReady,
				AppConditionHorizontalPodAutoscalerReady,
			},
		},
		"out of sync": {
			Init: func(status *AppStatus) {
				status.MarkSpaceHealthy()
				status.PropagateBuildStatus(happyBuild())
				status.PropagateEnvVarSecretStatus(envVarSecret())
				// out of sync
				status.PropagateDeploymentStatus(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "some-service-name",
						Generation: 42,
					},
				})
			},
			ExpectSucceeded: []apis.ConditionType{
				AppConditionSpaceReady,
				AppConditionBuildReady,
				AppConditionEnvVarSecretReady,
			},
			ExpectOngoing: []apis.ConditionType{
				AppConditionReady,
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
		binding     ServiceInstanceBinding
		expected    apis.ConditionType
		expectedErr error
	}{
		"correct": {
			binding: ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-binding",
				},
				Spec: ServiceInstanceBindingSpec{
					BindingType: BindingType{
						App: &AppRef{
							Name: "my-app",
						},
					},
					InstanceRef: corev1.LocalObjectReference{
						Name: "my-service-instance",
					},
				},
				Status: ServiceInstanceBindingStatus{
					BindingName: "my-service-instance",
				},
			},
			expected: apis.ConditionType("my-service-instanceReady"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := serviceBindingConditionType(tc.binding)
			testutil.AssertEqual(t, "conditionType", tc.expected, actual)
		})
	}
}

func TestAppStatus_PropagateDeploymentStatus(t *testing.T) {
	cases := map[string]struct {
		deployment    appsv1.Deployment
		wantCondition apis.Condition
	}{
		"generation mismatch": {
			deployment: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 300,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "GenerationOutOfDate",
				Message: "waiting for deployment spec update to be observed",
			},
		},
		"failed": {
			deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:    appsv1.DeploymentReplicaFailure,
							Status:  corev1.ConditionTrue,
							Reason:  "OOM",
							Message: "Out of memory",
						},
					},
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionFalse,
				Reason:  "OOM",
				Message: "Out of memory",
			},
		},
		"rolling out": {
			deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Replicas:            1,
					ReadyReplicas:       1,
					UpdatedReplicas:     1,
					UnavailableReplicas: 2,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "InitializingPods",
				Message: `waiting for deployment "" rollout to finish: 0 of 1 updated replicas are available`,
			},
		},
		"upgrade in-place": {
			deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Replicas:          2,
					ReadyReplicas:     2,
					UpdatedReplicas:   1,
					AvailableReplicas: 1,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "TerminatingOldReplicas",
				Message: `waiting for deployment "" rollout to finish: 1 old replicas are pending termination`,
			},
		},
		"starting up new pods": {
			deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Replicas:          2,
					ReadyReplicas:     1,
					UpdatedReplicas:   2,
					AvailableReplicas: 1,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "InitializingPods",
				Message: `waiting for deployment "" rollout to finish: 1 of 2 updated replicas are available`,
			},
		},
		"waiting for health checks": {
			deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Replicas:          2,
					ReadyReplicas:     2,
					UpdatedReplicas:   2,
					AvailableReplicas: 1,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionDeploymentReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "InitializingPods",
				Message: `waiting for deployment "" rollout to finish: 1 of 2 updated replicas are available`,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := AppStatus{}
			status.PropagateDeploymentStatus(&tc.deployment)

			actualCond := status.GetCondition(AppConditionDeploymentReady)

			testutil.AssertEqual(t, "condition type", tc.wantCondition.Type, actualCond.Type)
			testutil.AssertEqual(t, "condition status", tc.wantCondition.Status, actualCond.Status)
			testutil.AssertEqual(t, "condition reason", tc.wantCondition.Reason, actualCond.Reason)
			testutil.AssertEqual(t, "condition message", tc.wantCondition.Message, actualCond.Message)
		})
	}
}

func TestAppStatus_PropagateAutoscalerStatus(t *testing.T) {
	cases := map[string]struct {
		autoscaler    *autoscalingv1.HorizontalPodAutoscaler
		wantCondition apis.Condition
	}{
		"scaling up": {
			autoscaler: &autoscalingv1.HorizontalPodAutoscaler{
				Status: autoscalingv1.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 1,
					DesiredReplicas: 2,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionHorizontalPodAutoscalerReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "ScalingUp",
				Message: `waiting for autoscaler to finish scaling up: current replicas 1, target replicas 2`,
			},
		},
		"scaling down": {
			autoscaler: &autoscalingv1.HorizontalPodAutoscaler{
				Status: autoscalingv1.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 2,
					DesiredReplicas: 1,
				},
			},
			wantCondition: apis.Condition{
				Type:    AppConditionHorizontalPodAutoscalerReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "ScalingDown",
				Message: `waiting for autoscaler to finish scaling down: current replicas 2, target replicas 1`,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := AppStatus{}
			status.PropagateAutoscalerV1Status(tc.autoscaler)

			actualCond := status.GetCondition(AppConditionHorizontalPodAutoscalerReady)

			testutil.AssertEqual(t, "condition type", tc.wantCondition.Type, actualCond.Type)
			testutil.AssertEqual(t, "condition status", tc.wantCondition.Status, actualCond.Status)
			testutil.AssertEqual(t, "condition reason", tc.wantCondition.Reason, actualCond.Reason)
			testutil.AssertEqual(t, "condition message", tc.wantCondition.Message, actualCond.Message)
		})
	}
}
func TestAppStatus_PropagateRouteStatus(t *testing.T) {
	t.Parallel()

	buildQRB := func(name string) QualifiedRouteBinding {
		return QualifiedRouteBinding{
			Source: RouteSpecFields{
				Hostname: name,
				Domain:   "example.com",
				Path:     "/",
			},
			Destination: RouteDestination{
				Port:        DefaultRouteDestinationPort,
				ServiceName: name,
				Weight:      defaultRouteWeight,
			},
		}
	}

	buildRoute := func(qrb QualifiedRouteBinding) (route Route) {
		route.Name = "some-route"
		route.Generation = 1
		route.Status.ObservedGeneration = route.Generation
		route.Spec.RouteSpecFields = qrb.Source
		route.Status.InitializeConditions()
		route.Status.PropagateRouteSpecFields(qrb.Source)
		route.Status.PropagateVirtualService(&networking.VirtualService{
			ObjectMeta: metav1.ObjectMeta{Name: "some-vs"},
		}, nil, true)
		route.Status.PropagateBindings([]RouteDestination{qrb.Destination})
		route.Status.PropagateSpaceDomain(&SpaceDomain{
			Domain: qrb.Source.Domain,
		})
		route.Status.PropagateRouteServiceBinding(nil)

		// sanity check
		if !route.Status.IsReady() {
			t.Logf("%#v", route.Status.GetConditions())
			t.Fatal("Expected Route to be ready")
		}

		return
	}

	buildDesiredStatus := func(qrb QualifiedRouteBinding, isReady bool) (s AppRouteStatus) {
		s.QualifiedRouteBinding = qrb
		s.Status = RouteBindingStatusUnknown

		if isReady {
			s.VirtualService = corev1.LocalObjectReference{
				Name: "some-vs",
			}
			s.Status = RouteBindingStatusReady
		}
		s.URL = qrb.Source.String()
		return
	}

	buildOrphanedStatus := func(qrb QualifiedRouteBinding) (s AppRouteStatus) {
		s = buildDesiredStatus(qrb, false)
		s.Status = RouteBindingStatusOrphaned
		return s
	}

	bindingA := buildQRB("app-a")
	bindingB := buildQRB("app-b")
	routeA := buildRoute(bindingA)
	routeB := buildRoute(bindingB)

	cases := map[string]struct {
		bindings          []QualifiedRouteBinding
		routes            []Route
		extraBindings     []QualifiedRouteBinding
		wantCondition     apis.Condition
		wantRouteStatuses []AppRouteStatus
		wantURLs          []string
	}{
		"multiple success": {
			bindings: []QualifiedRouteBinding{bindingB, bindingA},
			routes:   []Route{routeB, routeA},
			wantCondition: apis.Condition{
				Type:   AppConditionRouteReady,
				Status: corev1.ConditionTrue,
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, true),
				buildDesiredStatus(bindingB, true),
			},
			wantURLs: []string{bindingA.Source.String(), bindingB.Source.String()},
		},
		"route missing": {
			bindings: []QualifiedRouteBinding{bindingA},
			routes:   []Route{routeB},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionFalse,
				Reason:  "RouteMissing",
				Message: "No Route defined for URL: app-a.example.com",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, false),
			},
			wantURLs: []string{bindingA.Source.String()},
		},
		"generation mismatch": {
			bindings: []QualifiedRouteBinding{bindingA},
			routes: []Route{func() Route {
				tmp := routeA.DeepCopy()
				tmp.Status.ObservedGeneration++
				return *tmp
			}()},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "RouteReconciling",
				Message: "The Route is currently updating",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, false),
			},
			wantURLs: []string{bindingA.Source.String()},
		},
		"unknown status": {
			bindings: []QualifiedRouteBinding{bindingA},
			routes: []Route{func() Route {
				tmp := routeA.DeepCopy()
				tmp.Status.manage().SetCondition(apis.Condition{
					Type:    RouteConditionReady,
					Status:  corev1.ConditionUnknown,
					Message: "some-message",
					Reason:  "SomeReason",
				})
				return *tmp
			}()},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "RouteUnhealthy",
				Message: "Route has status SomeReason: some-message",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, false),
			},
			wantURLs: []string{bindingA.Source.String()},
		},
		"failed route status": {
			bindings: []QualifiedRouteBinding{bindingA},
			routes: []Route{func() Route {
				tmp := routeA.DeepCopy()
				tmp.Status.manage().SetCondition(apis.Condition{
					Type:    RouteConditionReady,
					Status:  corev1.ConditionFalse,
					Message: "some-message",
					Reason:  "SomeReason",
				})
				return *tmp
			}()},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionFalse,
				Reason:  "RouteUnhealthy",
				Message: "Route has status SomeReason: some-message",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, false),
			},
			wantURLs: []string{bindingA.Source.String()},
		},
		"binding missing": {
			bindings: []QualifiedRouteBinding{bindingA},
			routes: []Route{func() Route {
				tmp := routeA.DeepCopy()
				tmp.Status.Bindings = nil
				return *tmp
			}()},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "RouteBindingPropagating",
				Message: "The binding is still propagating to the Route",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, false),
			},
			wantURLs: []string{bindingA.Source.String()},
		},
		"extra bindings": {
			bindings:      []QualifiedRouteBinding{bindingA},
			routes:        []Route{routeA},
			extraBindings: []QualifiedRouteBinding{bindingB},
			wantCondition: apis.Condition{
				Type:    AppConditionRouteReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "ExtraRouteBinding",
				Message: "The Route app-b.example.com has an extra binding to this App",
			},
			wantRouteStatuses: []AppRouteStatus{
				buildDesiredStatus(bindingA, true),
				buildOrphanedStatus(bindingB),
			},
			wantURLs: []string{bindingA.Source.String(), bindingB.Source.String()},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &AppStatus{}
			status.InitializeConditions()
			status.PropagateRouteStatus(tc.bindings, tc.routes, tc.extraBindings)

			var urls []string
			for _, r := range status.Routes {
				urls = append(urls, r.URL)
			}

			actualCond := status.GetCondition(AppConditionRouteReady)
			testutil.AssertEqual(t, "condition type", tc.wantCondition.Type, actualCond.Type)
			testutil.AssertEqual(t, "condition status", tc.wantCondition.Status, actualCond.Status)
			testutil.AssertEqual(t, "condition reason", tc.wantCondition.Reason, actualCond.Reason)
			testutil.AssertEqual(t, "condition message", tc.wantCondition.Message, actualCond.Message)
			testutil.AssertEqual(t, "route statuses", tc.wantRouteStatuses, status.Routes)
			testutil.AssertEqual(t, "URLs", tc.wantURLs, urls)
			testutil.AssertEqual(t, "Outer URLs", tc.wantURLs, status.URLs)
		})
	}
}

func TestAppInstancesStatus(t *testing.T) {
	t.Parallel()
	autoscalingRule := AppAutoscalingRule{
		RuleType: CPURuleType,
		Target:   ptr.Int32(80),
	}

	cases := map[string]struct {
		app     *App
		hpa     *autoscalingv1.HorizontalPodAutoscaler
		current InstanceStatus
		want    InstanceStatus
	}{
		"hpa is nil.": {
			current: InstanceStatus{},
			want:    InstanceStatus{},
		},
		"Propogated status successfully.": {
			app: &App{
				Spec: AppSpec{
					Instances: AppSpecInstances{
						Autoscaling: AppSpecAutoscaling{
							Rules: []AppAutoscalingRule{
								autoscalingRule,
							},
						},
					},
				},
			},
			hpa: &autoscalingv1.HorizontalPodAutoscaler{
				Status: autoscalingv1.HorizontalPodAutoscalerStatus{
					CurrentCPUUtilizationPercentage: ptr.Int32(70),
				},
			},
			current: InstanceStatus{},
			want: InstanceStatus{
				AutoscalingStatus: []AutoscalingRuleStatus{
					{
						AppAutoscalingRule: autoscalingRule,
						Current: AutoscalingRuleMetricValueStatus{
							AverageValue: resource.NewQuantity(70, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.current.PropagateAutoscalingStatus(tc.app, tc.hpa)
			testutil.AssertEqual(t, "AutoscalingStatus", tc.want, tc.current)
		})
	}
}

func TestPropagateADXBuildStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		u := dynamicutils.NewUnstructured(map[string]interface{}{
			"metadata.name": "some-name",
			"status.image":  "some-image",
		})

		s := new(AppStatus)
		err := s.PropagateADXBuildStatus(u)
		testutil.AssertErrorsEqual(t, nil, err)
	})
	t.Run("fail", func(t *testing.T) {
		u := dynamicutils.NewUnstructured(map[string]interface{}{
			"metadata.name": "some-name",
			"status.image":  []string{"wrong-type"},
		})
		s := new(AppStatus)
		err := s.PropagateADXBuildStatus(u)
		testutil.AssertErrorsEqual(t, errors.New("failed to read image from status: .status.image accessor error: [wrong-type] is of the type []string, expected string"), err)
	})
}

func TestAppStatus_PropagateVolumeBindingsStatus(t *testing.T) {
	t.Parallel()

	bindingA := ServiceInstanceBinding{
		Status: ServiceInstanceBindingStatus{
			VolumeStatus: &BindingVolumeStatus{
				PersistentVolumeName:      "pv-a",
				Mount:                     "/pv/a",
				PersistentVolumeClaimName: "pvc-a",
				ReadOnly:                  false,
				UidGid: UidGid{
					UID: "usr-a",
					GID: "grp-a",
				},
			},
		},
	}
	statusA := AppVolumeStatus{
		MountPath:       "/pv/a",
		VolumeName:      "pv-a",
		VolumeClaimName: "pvc-a",
		ReadOnly:        false,
		UidGid: UidGid{
			UID: "usr-a",
			GID: "grp-a",
		},
	}
	bindingB := ServiceInstanceBinding{
		Status: ServiceInstanceBindingStatus{
			VolumeStatus: &BindingVolumeStatus{
				PersistentVolumeName:      "pv-b",
				Mount:                     "/pv/b",
				PersistentVolumeClaimName: "pvc-b",
				ReadOnly:                  true,
				UidGid: UidGid{
					UID: "usr-b",
					GID: "grp-b",
				},
			},
		},
	}
	statusB := AppVolumeStatus{
		MountPath:       "/pv/b",
		VolumeName:      "pv-b",
		VolumeClaimName: "pvc-b",
		ReadOnly:        true,
		UidGid: UidGid{
			UID: "usr-b",
			GID: "grp-b",
		},
	}

	cases := map[string]struct {
		volumeBindings []*ServiceInstanceBinding
		expected       *AppStatus
	}{
		"no volume list": {
			volumeBindings: nil,
			expected: &AppStatus{
				Volumes: []AppVolumeStatus{},
			},
		},
		"empty volume list": {
			volumeBindings: []*ServiceInstanceBinding{},
			expected: &AppStatus{
				Volumes: []AppVolumeStatus{},
			},
		},
		"single volume": {
			volumeBindings: []*ServiceInstanceBinding{&bindingA},
			expected: &AppStatus{
				Volumes: []AppVolumeStatus{statusA},
			},
		},
		"sorted volumes pass through": {
			volumeBindings: []*ServiceInstanceBinding{&bindingA, &bindingB},
			expected: &AppStatus{
				Volumes: []AppVolumeStatus{statusA, statusB},
			},
		},
		"reverse sorted volumes get sorted": {
			volumeBindings: []*ServiceInstanceBinding{&bindingB, &bindingA},
			expected: &AppStatus{
				Volumes: []AppVolumeStatus{statusA, statusB},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := new(AppStatus)
			actual.PropagateVolumeBindingsStatus(tc.volumeBindings)

			testutil.AssertEqual(t, "status", tc.expected, actual)
		})
	}
}

func TestAppStatus_PropagateServiceInstanceBindingsStatus(t *testing.T) {
	t.Parallel()

	bindingA := ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding-a",
		},
		Status: ServiceInstanceBindingStatus{
			BindingName: "binding-a-name",
		},
	}
	bindingB := ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding-b",
		},
		Status: ServiceInstanceBindingStatus{
			BindingName: "binding-b-name",
		},
	}

	cases := map[string]struct {
		bindings                    []ServiceInstanceBinding
		expectedServiceBindingNames []string
	}{
		"no binding list": {
			bindings:                    nil,
			expectedServiceBindingNames: []string{},
		},
		"empty binding list": {
			bindings:                    []ServiceInstanceBinding{},
			expectedServiceBindingNames: []string{},
		},
		"single binding": {
			bindings:                    []ServiceInstanceBinding{bindingA},
			expectedServiceBindingNames: []string{"binding-a-name"},
		},
		"sorted volumes pass through": {
			bindings:                    []ServiceInstanceBinding{bindingA, bindingB},
			expectedServiceBindingNames: []string{"binding-a-name", "binding-b-name"},
		},
		"reverse sorted volumes get sorted": {
			bindings:                    []ServiceInstanceBinding{bindingB, bindingA},
			expectedServiceBindingNames: []string{"binding-a-name", "binding-b-name"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := new(AppStatus)
			actual.PropagateServiceInstanceBindingsStatus(tc.bindings)

			testutil.AssertEqual(t, "status", tc.expectedServiceBindingNames, actual.ServiceBindingNames)
		})
	}
}
