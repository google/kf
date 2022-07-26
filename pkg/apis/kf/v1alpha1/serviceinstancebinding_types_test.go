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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestBindingOSBStatus_IsBlank(t *testing.T) {
	t.Parallel()

	statePtr := &OSBState{}

	cases := map[string]struct {
		status    BindingOSBStatus
		wantBlank bool
	}{
		"blank": {
			wantBlank: true,
		},
		"Binding": {
			status:    BindingOSBStatus{Binding: statePtr},
			wantBlank: false,
		},
		"Bound": {
			status:    BindingOSBStatus{Bound: statePtr},
			wantBlank: false,
		},
		"BindFailed": {
			status:    BindingOSBStatus{BindFailed: statePtr},
			wantBlank: false,
		},
		"Unbinding": {
			status:    BindingOSBStatus{Unbinding: statePtr},
			wantBlank: false,
		},
		"Unbound": {
			status:    BindingOSBStatus{Unbound: statePtr},
			wantBlank: false,
		},
		"UnbindFailed": {
			status:    BindingOSBStatus{UnbindFailed: statePtr},
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
