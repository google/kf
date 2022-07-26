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

package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/ptr"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

func TestServiceInstanceStatus_PropagateSecretStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		secret              *corev1.Secret
		wantSecretReady     corev1.ConditionStatus
		wantSecretPopulated corev1.ConditionStatus
	}{
		"no secret": {
			secret:              nil,
			wantSecretReady:     corev1.ConditionUnknown,
			wantSecretPopulated: corev1.ConditionUnknown,
		},
		"nil secret data": {
			secret: &corev1.Secret{
				Data: nil,
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionUnknown,
		},
		"blank secret data": {
			secret: &corev1.Secret{
				Data: make(map[string][]byte),
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionUnknown,
		},
		"secret without correct key": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"some-key": []byte("some-value"),
				},
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionUnknown,
		},
		"bad populated secret": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceParamsSecretKey: []byte("this-is-not-json"),
				},
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionFalse,
		},
		"blank JSON secret": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceParamsSecretKey: []byte("{}"),
				},
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionTrue,
		},
		"non-blank JSON secret": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceParamsSecretKey: []byte(`{"foo":"bar"}`),
				},
			},
			wantSecretReady:     corev1.ConditionTrue,
			wantSecretPopulated: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceStatus{}
			status.InitializeConditions()

			status.PropagateSecretStatus(tc.secret)

			actualSecretReady := status.manage().GetCondition(ServiceInstanceConditionParamsSecretReady)
			actualSecretPopulated := status.manage().GetCondition(ServiceInstanceConditionParamsSecretPopulatedReady)

			testutil.AssertEqual(t, "secret ready condition", tc.wantSecretReady, actualSecretReady.Status)
			testutil.AssertEqual(t, "secret populated condition", tc.wantSecretPopulated, actualSecretPopulated.Status)
		})
	}
}

func TestServiceInstanceStatus_PropagateDeprovisionStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.DeprovisionResponse
		err           error
		wantCondition corev1.ConditionStatus
	}{
		"500 error fails": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 500},
			wantCondition: corev1.ConditionFalse,
		},
		"409 error fails": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 409},
			wantCondition: corev1.ConditionFalse,
		},
		"410 error deletes": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 410},
			wantCondition: corev1.ConditionTrue,
		},
		"404 error deletes": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 404},
			wantCondition: corev1.ConditionTrue,
		},
		"other error fails": {
			err:           errors.New("other"),
			wantCondition: corev1.ConditionFalse,
		},
		"async operation continues": {
			response:      &osbclient.DeprovisionResponse{Async: true},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.DeprovisionResponse{},
			wantCondition: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceStatus{}
			status.InitializeConditions()
			status.OSBStatus = OSBStatus{
				Provisioned: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateDeprovisionStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertDeletionInvariant(t, original, status)
		})
	}
}

func TestServiceInstanceStatus_PropagateDeprovisionAsyncStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.LastOperationResponse
		err           error
		wantCondition corev1.ConditionStatus
	}{
		"500 error retries": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 500},
			wantCondition: corev1.ConditionUnknown,
		},
		"409 error retries": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 409},
			wantCondition: corev1.ConditionUnknown,
		},
		"410 error deletes": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 410},
			wantCondition: corev1.ConditionTrue,
		},
		"404 error deletes": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 404},
			wantCondition: corev1.ConditionTrue,
		},
		"other error fails": {
			err:           errors.New("other"),
			wantCondition: corev1.ConditionFalse,
		},
		"in-progress operation continues": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateInProgress},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateSucceeded},
			wantCondition: corev1.ConditionTrue,
		},
		"failed operation completes": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateFailed},
			wantCondition: corev1.ConditionFalse,
		},
		"unknown operation fails": {
			response:      &osbclient.LastOperationResponse{State: "badstate"},
			wantCondition: corev1.ConditionFalse,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceStatus{}
			status.InitializeConditions()
			status.OSBStatus = OSBStatus{
				Deprovisioning: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateDeprovisionAsyncStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertDeletionInvariant(t, original, status)
		})
	}
}

// assertDeletionInvariant checks the invariant that condition and status are
// set and that they're in a valid state.
func assertDeletionInvariant(t *testing.T, original, updated *ServiceInstanceStatus) {
	actualCondition := updated.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)

	actualOSB := updated.OSBStatus

	switch actualCondition.Status {
	case corev1.ConditionTrue:
		testutil.AssertTrue(t, "OSBStatus deprovisioned", actualOSB.Deprovisioned != nil)
	case corev1.ConditionFalse:
		testutil.AssertTrue(t, "OSBStatus deprovisionFailed", actualOSB.DeprovisionFailed != nil)
	case corev1.ConditionUnknown:
		if actualOSB.Deprovisioning == nil && !reflect.DeepEqual(actualOSB, original.OSBStatus) {
			t.Errorf("expected deprovisioning or no change got: %#v", actualOSB)
		}
	default:
		t.Fatal("expected condition to be set")
	}
}

func TestServiceInstanceStatus_PropagateProvisionStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.ProvisionResponse
		err           error
		wantCondition corev1.ConditionStatus
	}{
		"500 error fails": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 500},
			wantCondition: corev1.ConditionFalse,
		},
		"409 error fails": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 409},
			wantCondition: corev1.ConditionFalse,
		},
		"other error fails": {
			err:           errors.New("other"),
			wantCondition: corev1.ConditionFalse,
		},
		"async operation continues": {
			response:      &osbclient.ProvisionResponse{Async: true},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.ProvisionResponse{},
			wantCondition: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceStatus{}
			status.InitializeConditions()
			status.OSBStatus = OSBStatus{
				Provisioned: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateProvisionStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertCreationInvariant(t, original, status)
		})
	}
}

func TestServiceInstanceStatus_PropagateProvisionAsyncStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.LastOperationResponse
		err           error
		wantCondition corev1.ConditionStatus
	}{
		"500 error retries": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 500},
			wantCondition: corev1.ConditionUnknown,
		},
		"409 error retries": {
			err:           &osbclient.HTTPStatusCodeError{StatusCode: 409},
			wantCondition: corev1.ConditionUnknown,
		},
		"other error fails": {
			err:           errors.New("other"),
			wantCondition: corev1.ConditionFalse,
		},
		"in-progress operation continues": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateInProgress},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateSucceeded},
			wantCondition: corev1.ConditionTrue,
		},
		"failed operation completes": {
			response:      &osbclient.LastOperationResponse{State: osbclient.StateFailed},
			wantCondition: corev1.ConditionFalse,
		},
		"unknown operation fails": {
			response:      &osbclient.LastOperationResponse{State: "badstate"},
			wantCondition: corev1.ConditionFalse,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceStatus{}
			status.InitializeConditions()
			status.OSBStatus = OSBStatus{
				Deprovisioning: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateProvisionAsyncStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertCreationInvariant(t, original, status)
		})
	}
}

// assertCreationInvariant checks the invariant that condition and status are
// set and that they're in a valid state.
func assertCreationInvariant(t *testing.T, original, updated *ServiceInstanceStatus) {
	actualCondition := updated.manage().GetCondition(ServiceInstanceConditionBackingResourceReady)

	actualOSB := updated.OSBStatus

	switch actualCondition.Status {
	case corev1.ConditionTrue:
		testutil.AssertTrue(t, "OSBStatus provisioned", actualOSB.Provisioned != nil)
	case corev1.ConditionFalse:
		testutil.AssertTrue(t, "OSBStatus provisionFailed", actualOSB.ProvisionFailed != nil)
	case corev1.ConditionUnknown:
		if actualOSB.Provisioning == nil && !reflect.DeepEqual(actualOSB, original.OSBStatus) {
			t.Errorf("expected provisioning or no change got: %#v", actualOSB)
		}
	default:
		t.Fatal("expected condition to be set")
	}
}

func TestFormatOperationMessage(t *testing.T) {

	cases := map[string]struct {
		response *osbclient.LastOperationResponse
		want     string
	}{
		"nil operation": {
			want: "(nil operation)",
		},
		"no description operation": {
			response: &osbclient.LastOperationResponse{
				State: osbclient.StateInProgress,
			},
			want: `operation state: "in progress"`,
		},
		"full operation": {
			response: &osbclient.LastOperationResponse{
				State:       osbclient.StateInProgress,
				Description: ptr.String("some info here"),
			},
			want: `operation state: "in progress" description: "some info here"`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "output", tc.want, formatOperationMessage(tc.response))
		})
	}

}
