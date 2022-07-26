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
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestOSBStatus_IsBlank(t *testing.T) {
	t.Parallel()

	statePtr := &OSBState{}

	cases := map[string]struct {
		status    OSBStatus
		wantBlank bool
	}{
		"blank": {
			wantBlank: true,
		},
		"Provisioning": {
			status:    OSBStatus{Provisioning: statePtr},
			wantBlank: false,
		},
		"Provisioned": {
			status:    OSBStatus{Provisioned: statePtr},
			wantBlank: false,
		},
		"ProvisionFailed": {
			status:    OSBStatus{ProvisionFailed: statePtr},
			wantBlank: false,
		},
		"Deprovisioning": {
			status:    OSBStatus{Deprovisioning: statePtr},
			wantBlank: false,
		},
		"Deprovisioned": {
			status:    OSBStatus{Deprovisioned: statePtr},
			wantBlank: false,
		},
		"DeprovisionFailed": {
			status:    OSBStatus{DeprovisionFailed: statePtr},
			wantBlank: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotBlank := tc.status.IsBlank()
			testutil.AssertEqual(t, "blank", tc.wantBlank, gotBlank)
		})
	}
}

func TestParseVolumeInstanceParams(t *testing.T) {
	share, capacity := "192.168.0.1/test", "1Gi"
	goodSecret := &corev1.Secret{}
	goodSecret.Data = map[string][]byte{
		ServiceInstanceParamsSecretKey: []byte(`{"share":"192.168.0.1/test", "capacity":"1Gi"}`),
	}

	partialSecret := &corev1.Secret{}
	partialSecret.Data = map[string][]byte{
		ServiceInstanceParamsSecretKey: []byte(`{"share":"192.168.0.1/test"}`),
	}

	secretWithUnexpectedField := new(corev1.Secret)
	secretWithUnexpectedField.Data = map[string][]byte{
		ServiceInstanceParamsSecretKey: []byte(`{"extra":"extra","share":"192.168.0.1/test", "capacity":"1Gi"}`),
	}

	cases := map[string]struct {
		secret *corev1.Secret
		want   *VolumeInstanceParams
		err    error
	}{
		"Secret no params": {
			secret: &corev1.Secret{},
			want:   nil,
			err:    fmt.Errorf("secret is missing key %q", ServiceInstanceParamsSecretKey),
		},
		"Secret with extra data": {
			secret: goodSecret,
			want: &VolumeInstanceParams{
				Share:    share,
				Capacity: capacity,
			},
		},
		"Good secret": {
			secret: goodSecret,
			want: &VolumeInstanceParams{
				Share:    share,
				Capacity: capacity,
			},
		},
		"Partial secret": {
			secret: partialSecret,
			want: &VolumeInstanceParams{
				Share: share,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			params, err := ParseVolumeInstanceParams(tc.secret)
			if tc.err != nil {
				testutil.AssertEqual(t, "err", tc.err, err)
				return
			}
			testutil.AssertNil(t, "err", err)
			testutil.AssertEqual(t, "params", *params, *tc.want)
		})
	}
}
