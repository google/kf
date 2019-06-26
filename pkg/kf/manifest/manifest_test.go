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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/manifest"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewFromReader(t *testing.T) {
	cases := map[string]struct {
		fileContent string
		expected    *manifest.Manifest
	}{
		"app name": {
			fileContent: `---
applications:
- name: MY-APP
`,
			expected: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name: "MY-APP",
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := manifest.NewFromReader(strings.NewReader(tc.fileContent))
			testutil.AssertNil(t, "error", err)
			testutil.AssertEqual(t, "manifest", tc.expected, actual)
		})
	}
}

func TestCheckForManifest(t *testing.T) {
	cases := map[string]struct {
		fileName    string
		fileContent string
		expected    *manifest.Manifest
	}{
		"yml": {
			fileName: "manifest.yml",
			fileContent: `---
applications:
- name: MY-APP
`,
			expected: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name: "MY-APP",
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "kf-manifest-test")
			testutil.AssertNil(t, "error creating test directory", err)
			defer func() {
				testutil.AssertNil(t, "error deleting test directory", os.RemoveAll(dir))
			}()

			err = ioutil.WriteFile(filepath.Join(dir, tc.fileName), []byte(tc.fileContent), 0644)
			testutil.AssertNil(t, "error writing manifest file", err)

			actual, err := manifest.CheckForManifest(dir)
			testutil.AssertNil(t, "error", err)
			testutil.AssertEqual(t, "manifest", tc.expected, actual)
		})
	}
}
