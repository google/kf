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

package config

import (
	"net/http"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"k8s.io/client-go/transport"
)

func TestImpersonationRequest(t *testing.T) {

	cases := map[string]struct {
		params KfParams
	}{
		"no impersonation": {
			params: KfParams{},
		},
		"impersonate user": {
			params: KfParams{
				Impersonate: transport.ImpersonationConfig{
					UserName: "test@google.com",
				},
			},
		},
		"impersonate group": {
			params: KfParams{
				Impersonate: transport.ImpersonationConfig{
					Groups: []string{"system:masters"},
				},
			},
		},
		"impersonate extra": {
			params: KfParams{
				Impersonate: transport.ImpersonationConfig{
					Extra: map[string][]string{
						"extra-key": {"extra-value"},
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://example.com", nil)
			testutil.AssertNil(t, "error", err)

			recorder := &recordingTransport{}
			wrapper := NewImpersonatingRoundTripperWrapper(&tc.params)
			wrapper(recorder).RoundTrip(req)

			testutil.AssertGolden(t, "request", []byte(recorder.requestDump))
		})
	}
}
