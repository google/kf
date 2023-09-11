// Copyright 2023 Google LLC
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

package featureflag

import (
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestValidateKfNamespaceName(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		key  string
		want error
	}{
		"nominal":           {key: "kf/", want: nil},
		"invalid namespace": {key: "kf-system/", want: errors.New(`invalid namespace "kf-system" queued, expect only "kf"`)},
		"invalid key":       {key: "foo/bar/bazz", want: errors.New(`unexpected key format: "foo/bar/bazz"`)},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateKfNamespaceName(tc.key)

			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}
