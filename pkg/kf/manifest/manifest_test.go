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

package manifest_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/manifest"
)

func TestNewFromReader(t *testing.T) {

	cases := map[string]struct {
		fileContent string
		expected    manifest.Manifest
	}{
		"app name": {
			fileContent: `---
applications:
- name: MY-APP
`,
			expected: manifest.Manifest{
				Applications: []manifest.Application{
					manifest.Application{
						Name: "MY-APP",
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := manifest.NewFromReader(strings.NewReader(tc.fileContent))
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(*actual, tc.expected) {
				t.Error(fmt.Printf("not equal: %v, %v", tc.expected, actual))
			}
		})
	}
}
