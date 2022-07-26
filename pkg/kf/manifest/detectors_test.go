// Copyright 2020 Google LLC
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
// limitations under the License

package manifest

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestLooksLikeBuildpackV2URL(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected bool
	}{
		"unmapped buildpack string": {
			input:    "java_buildpack",
			expected: false,
		},
		"valid buildpack URL": {
			input:    "https://github.com/cloudfoundry/staticfile-buildpack",
			expected: true,
		},
		"buildpack URL with version": {
			input:    "https://github.com/cloudfoundry/staticfile-buildpack.git#v3.11.2",
			expected: true,
		},
		"possibly valid URL": {
			input:    "http://example.com/buildpack",
			expected: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "looks like URL", tc.expected, looksLikeBuildpackV2URL(tc.input))
		})
	}
}
