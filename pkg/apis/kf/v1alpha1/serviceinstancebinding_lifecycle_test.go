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
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

func TestServiceInstanceBindingStatus_PropagateSecretStatus(t *testing.T) {
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
			status := &ServiceInstanceBindingStatus{}
			status.InitializeConditions()

			status.PropagateParamsSecretStatus(tc.secret)

			actualSecretReady := status.manage().GetCondition(ServiceInstanceBindingConditionParamsSecretReady)
			actualSecretPopulated := status.manage().GetCondition(ServiceInstanceBindingConditionParamsSecretPopulatedReady)

			testutil.AssertEqual(t, "secret ready condition", tc.wantSecretReady, actualSecretReady.Status)
			testutil.AssertEqual(t, "secret populated condition", tc.wantSecretPopulated, actualSecretPopulated.Status)
		})
	}
}

func TestServiceInstanceBindingStatus_PropagateUnbindStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.UnbindResponse
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
			response:      &osbclient.UnbindResponse{Async: true},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.UnbindResponse{},
			wantCondition: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceBindingStatus{}
			status.InitializeConditions()
			status.OSBStatus = BindingOSBStatus{
				Bound: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateUnbindStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertUnbindInvariant(t, original, status)
		})
	}
}

func TestServiceInstanceBindingStatus_PropagateUnbindLastOperationStatus(t *testing.T) {
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
			status := &ServiceInstanceBindingStatus{}
			status.InitializeConditions()
			status.OSBStatus = BindingOSBStatus{
				Unbinding: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateUnbindLastOperationStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertUnbindInvariant(t, original, status)
		})
	}
}

// assertUnbindInvariant checks the invariant that condition and status are
// set and that they're in a valid state.
func assertUnbindInvariant(t *testing.T, original, updated *ServiceInstanceBindingStatus) {
	actualCondition := updated.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)

	actualOSB := updated.OSBStatus

	switch actualCondition.Status {
	case corev1.ConditionTrue:
		testutil.AssertTrue(t, "OSBStatus unbound", actualOSB.Unbound != nil)
	case corev1.ConditionFalse:
		testutil.AssertTrue(t, "OSBStatus unbindFailed", actualOSB.UnbindFailed != nil)
	case corev1.ConditionUnknown:
		if actualOSB.Unbinding == nil && !reflect.DeepEqual(actualOSB, original.OSBStatus) {
			t.Errorf("expected unbinding or no change got: %#v", actualOSB)
		}
	default:
		t.Fatal("expected condition to be set")
	}
}

func TestServiceInstanceBindingStatus_PropagateBindStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		response      *osbclient.BindResponse
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
			response:      &osbclient.BindResponse{Async: true},
			wantCondition: corev1.ConditionUnknown,
		},
		"successful operation completes": {
			response:      &osbclient.BindResponse{},
			wantCondition: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceBindingStatus{}
			status.InitializeConditions()
			original := status.DeepCopy()

			status.PropagateBindStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertBindInvariant(t, original, status)
		})
	}
}

func TestServiceInstanceBindingStatus_PropagateBindLastOperationStatus(t *testing.T) {
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
			status := &ServiceInstanceBindingStatus{}
			status.InitializeConditions()
			status.OSBStatus = BindingOSBStatus{
				Binding: &OSBState{},
			}
			original := status.DeepCopy()

			status.PropagateBindLastOperationStatus(tc.response, tc.err)

			actualCondition := status.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)

			assertBindInvariant(t, original, status)
		})
	}

}
func TestServiceInstanceBindingStatus_PropagateVolumeStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		secret        *corev1.Secret
		instance      *ServiceInstance
		wantStatus    *BindingVolumeStatus
		wantCondition corev1.ConditionStatus
	}{
		"Volume status not populated in serviceinstance": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"some-key": []byte("some-value"),
				},
			},
			instance:      &ServiceInstance{},
			wantStatus:    nil,
			wantCondition: corev1.ConditionTrue,
		},
		"no relevant field in secret": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"some-key": []byte("some-value"),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus:    nil,
			wantCondition: corev1.ConditionFalse,
		},
		"correct fields, gid and uid not specified": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"mount":"mount"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus: &BindingVolumeStatus{
				Mount:                     "mount",
				PersistentVolumeName:      "pv",
				PersistentVolumeClaimName: "pvc",
			},
			wantCondition: corev1.ConditionTrue,
		},
		"gid invalid": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"mount":"mount", "gid": "-1000"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus:    nil,
			wantCondition: corev1.ConditionFalse,
		},
		"uid invalid": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"mount":"mount", "uid": "-1000"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantCondition: corev1.ConditionFalse,
		},
		"uid not specified": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"mount":"mount", "gid": "2000"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus: &BindingVolumeStatus{
				Mount:                     "mount",
				PersistentVolumeName:      "pv",
				PersistentVolumeClaimName: "pvc",
				UidGid: UidGid{
					GID: "2000",
				},
			},
			wantCondition: corev1.ConditionTrue,
		},
		"mount missing": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"gid":"2000","uid":"2000"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus:    nil,
			wantCondition: corev1.ConditionFalse,
		},
		"all fields in secret": {
			secret: &corev1.Secret{
				Data: map[string][]byte{
					ServiceInstanceBindingParamsSecretKey: []byte(`{"mount":"mount","gid":"2000","uid":"2000"}`),
				},
			},
			instance: &ServiceInstance{
				Status: ServiceInstanceStatus{
					VolumeStatus: &VolumeStatus{
						PersistentVolumeName:      "pv",
						PersistentVolumeClaimName: "pvc",
					},
				},
			},
			wantStatus: &BindingVolumeStatus{
				Mount:                     "mount",
				PersistentVolumeName:      "pv",
				PersistentVolumeClaimName: "pvc",
				ReadOnly:                  false,
				UidGid: UidGid{
					UID: "2000",
					GID: "2000",
				},
			},
			wantCondition: corev1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &ServiceInstanceBindingStatus{}
			status.PropagateVolumeStatus(tc.instance, tc.secret)

			actualCondition := status.manage().GetCondition(ServiceInstanceBindingConditionVolumeParamsPopulatedReady)
			testutil.AssertEqual(t, "condition", tc.wantCondition, actualCondition.Status)
			testutil.AssertEqual(t, "condition", tc.wantStatus, status.VolumeStatus)
		})
	}
}

// assertBindInvariant checks the invariant that condition and status are
// set and that they're in a valid state.
func assertBindInvariant(t *testing.T, original, updated *ServiceInstanceBindingStatus) {
	actualCondition := updated.manage().GetCondition(ServiceInstanceBindingConditionBackingResourceReady)

	actualOSB := updated.OSBStatus

	switch actualCondition.Status {
	case corev1.ConditionTrue:
		testutil.AssertTrue(t, "OSBStatus bound", actualOSB.Bound != nil)
	case corev1.ConditionFalse:
		testutil.AssertTrue(t, "OSBStatus bindFailed", actualOSB.BindFailed != nil)
	case corev1.ConditionUnknown:
		if actualOSB.Binding == nil && !reflect.DeepEqual(actualOSB, original.OSBStatus) {
			t.Errorf("expected binding or no change got: %#v", actualOSB)
		}
	default:
		t.Fatal("expected condition to be set")
	}
}
