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
	"errors"
	"testing"

	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	apitesting "knative.dev/pkg/apis/testing"
)

func TestRoute_HappyPath(t *testing.T) {
	t.Parallel()

	status := RouteStatus{}
	status.InitializeConditions()

	// sanity check conditions get initiailized as unknown
	for _, cond := range status.Conditions {
		apitesting.CheckConditionOngoing(status.duck(), cond.Type, t)
	}

	// Check to see that each status we expect is pending
	apitesting.CheckConditionOngoing(status.duck(), RouteConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), RouteConditionVirtualServiceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), RouteConditionSpaceDomainReady, t)
	apitesting.CheckConditionOngoing(status.duck(), RouteConditionRouteServiceReady, t)

	// Mark expected necessary statuses as successful
	status.RouteServiceCondition().MarkSuccess()
	status.VirtualServiceCondition().MarkSuccess()
	status.SpaceDomainCondition().MarkSuccess()

	// Check to see that the overall status has succeeded
	apitesting.CheckConditionSucceeded(status.duck(), RouteConditionReady, t)
}

func TestRouteStatus_VirtualServiceError(t *testing.T) {
	t.Parallel()

	status := RouteStatus{}
	status.PropagateVirtualService(nil, errors.New("some-reason: some-message"))
	apitesting.CheckConditionFailed(status.duck(), RouteConditionVirtualServiceReady, t)
}

func TestRouteStatus_VirtualServiceNil(t *testing.T) {
	t.Parallel()

	status := RouteStatus{}
	status.PropagateVirtualService(nil, nil)
	testutil.AssertEqual(t, "VirtualService", "", status.VirtualService.Name)
}

func TestRouteStatus_VirtualService(t *testing.T) {
	t.Parallel()

	status := RouteStatus{}
	status.PropagateVirtualService(&networking.VirtualService{
		ObjectMeta: metav1.ObjectMeta{Name: "some-name"},
	}, nil)
	testutil.AssertEqual(t, "VirtualService", "some-name", status.VirtualService.Name)
}

func TestRouteStatus_RouteSpecFields(t *testing.T) {
	t.Parallel()

	status := RouteStatus{}
	status.PropagateRouteSpecFields(RouteSpecFields{
		Hostname: "some-host",
	})
	testutil.AssertEqual(t, "RouteSpecFields.Hostname", "some-host", status.RouteSpecFields.Hostname)
}

func TestRouteStatus_PropagateBindings(t *testing.T) {
	cases := map[string]struct {
		bindings []RouteDestination
		want     RouteStatus
	}{
		"bindings": {
			bindings: []RouteDestination{
				{
					ServiceName: "some-app",
					Port:        80,
					Weight:      3,
				},
			},
			want: RouteStatus{
				Bindings: []RouteDestination{
					{
						ServiceName: "some-app",
						Port:        80,
						Weight:      3,
					},
				},
				AppBindingDisplayNames: []string{
					"some-app",
				},
			},
		},
		"empty bindings": {
			bindings: []RouteDestination{},
			want:     RouteStatus{},
		},
		"nil bindings": {
			bindings: nil,
			want:     RouteStatus{},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := RouteStatus{}
			status.PropagateBindings(tc.bindings)

			testutil.AssertEqual(t, "RouteStatus", tc.want, status)
		})
	}
}

func TestRouteStatus_PropagateSpaceDomain(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		spaceDomain   *SpaceDomain
		wantCondition apis.Condition
	}{
		"nil domain": {
			spaceDomain: nil,
			wantCondition: apis.Condition{
				Type:    RouteConditionSpaceDomainReady,
				Status:  corev1.ConditionFalse,
				Reason:  "ReconciliationError",
				Message: "Error occurred while InvalidDomain SpaceDomain: The domain specified on the Route isn't permitted by the Space",
			},
		},
		"set domain": {
			spaceDomain: &SpaceDomain{Domain: "example.com"},
			wantCondition: apis.Condition{
				Type:   RouteConditionSpaceDomainReady,
				Status: corev1.ConditionTrue,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := RouteStatus{}
			status.PropagateSpaceDomain(tc.spaceDomain)

			gotCondition := status.GetCondition(RouteConditionSpaceDomainReady)
			testutil.AssertNotNil(t, "condition", gotCondition)
			gotCondition.LastTransitionTime = apis.VolatileTime{} // clear non-deterministic time
			testutil.AssertEqual(t, "condition", tc.wantCondition, *gotCondition)
		})
	}
}

func TestRouteStatus_PropagateRouteServiceBinding(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routeServices []RouteServiceDestination
		wantStatus    RouteStatus
		wantCondition apis.Condition
	}{
		"no route service": {
			routeServices: nil,
			wantStatus: RouteStatus{
				RouteService: corev1.LocalObjectReference{},
			},
			wantCondition: apis.Condition{
				Type:   RouteConditionRouteServiceReady,
				Status: corev1.ConditionTrue,
			},
		},
		"one route service": {
			routeServices: []RouteServiceDestination{
				{
					Name: "google",
					RouteServiceURL: &RouteServiceURL{
						Host: "google.com",
					},
				},
			},
			wantStatus: RouteStatus{
				RouteService: corev1.LocalObjectReference{
					Name: "google",
				},
			},
			wantCondition: apis.Condition{
				Type:   RouteConditionRouteServiceReady,
				Status: corev1.ConditionTrue,
			},
		},
		"multiple route services": {
			routeServices: []RouteServiceDestination{
				{
					Name: "google",
					RouteServiceURL: &RouteServiceURL{
						Host: "google.com",
					},
				},
				{
					Name: "yahoo",
					RouteServiceURL: &RouteServiceURL{
						Host: "yahoo.com",
					},
				},
			},
			wantStatus: RouteStatus{
				RouteService: corev1.LocalObjectReference{
					Name: "yahoo",
				},
			},
			wantCondition: apis.Condition{
				Type:    RouteConditionRouteServiceReady,
				Status:  corev1.ConditionFalse,
				Reason:  "MultipleRouteServices",
				Message: "More than one route service is bound: [google, yahoo]",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := RouteStatus{}
			status.PropagateRouteServiceBinding(tc.routeServices)

			gotCondition := status.GetCondition(RouteConditionRouteServiceReady)
			testutil.AssertNotNil(t, "condition", gotCondition)
			gotCondition.LastTransitionTime = apis.VolatileTime{} // clear non-deterministic time
			testutil.AssertEqual(t, "condition", tc.wantCondition, *gotCondition)
		})
	}
}
