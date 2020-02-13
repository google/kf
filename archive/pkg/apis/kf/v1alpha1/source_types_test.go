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
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
)

func TestSourceSpec_NeedsUpdateRequestsIncrement(t *testing.T) {
	current := SourceSpec{
		UpdateRequests: 33,
		ContainerImage: SourceSpecContainerImage{
			Image: "mysql/mysql",
		},
	}

	cases := map[string]struct {
		old                  SourceSpec
		expectNeedsIncrement bool
	}{
		"matches exactly": {
			old: SourceSpec{
				UpdateRequests: 33,
				ContainerImage: SourceSpecContainerImage{
					Image: "mysql/mysql",
				},
			},
			expectNeedsIncrement: false,
		},
		"increment already done": {
			old: SourceSpec{
				UpdateRequests: 34,
				ContainerImage: SourceSpecContainerImage{
					Image: "mariadb/mariadb",
				},
			},
			expectNeedsIncrement: false,
		},
		"needs increment": {
			old: SourceSpec{
				UpdateRequests: 33,
				ContainerImage: SourceSpecContainerImage{
					Image: "mariadb/mariadb",
				},
			},
			expectNeedsIncrement: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := current.NeedsUpdateRequestsIncrement(tc.old)

			testutil.AssertEqual(t, "needsIncrement", tc.expectNeedsIncrement, actual)
		})
	}
}
