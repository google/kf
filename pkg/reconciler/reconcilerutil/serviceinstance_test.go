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

package reconcilerutil

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

func TestIsConflictOSBError(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		err  error
		want bool
	}{
		"409 error true": {
			err:  &osbclient.HTTPStatusCodeError{StatusCode: 409},
			want: true,
		},
		"200 error false": {
			err:  &osbclient.HTTPStatusCodeError{StatusCode: 200},
			want: false,
		},
		"500 error false": {
			err:  &osbclient.HTTPStatusCodeError{StatusCode: 500},
			want: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := IsConflictOSBError(tc.err)

			testutil.AssertEqual(t, "conflict error", tc.want, got)
		})
	}
}
