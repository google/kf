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
		"buildpacks": {
			fileContent: `---
applications:
- name: MY-APP
  stack: cflinuxfs3
  buildpacks:
  - java
  - node
`,
			expected: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:       "MY-APP",
						Stack:      "cflinuxfs3",
						Buildpacks: []string{"java", "node"},
					},
				},
			},
		},
		"docker": {
			fileContent: `---
applications:
- name: MY-APP
  docker:
    image: "gcr.io/my-image"
`,
			expected: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name: "MY-APP",
						Docker: manifest.AppDockerImage{
							Image: "gcr.io/my-image",
						},
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

func TestOverride(t *testing.T) {
	cases := map[string]struct {
		base     manifest.Application
		override manifest.Application
		expected manifest.Application
	}{
		"blank": {},
		"basic fields, blank override": {
			base:     manifest.Application{Name: "hello"},
			override: manifest.Application{},
			expected: manifest.Application{Name: "hello"},
		},
		"nested field, blank override": {
			base:     manifest.Application{Docker: manifest.AppDockerImage{Image: "hello"}},
			override: manifest.Application{},
			expected: manifest.Application{Docker: manifest.AppDockerImage{Image: "hello"}},
		},
		"basic fields override": {
			base:     manifest.Application{Name: "hello"},
			override: manifest.Application{Name: "override"},
			expected: manifest.Application{Name: "override"},
		},
		"nested field override": {
			base:     manifest.Application{Docker: manifest.AppDockerImage{Image: "hello"}},
			override: manifest.Application{Docker: manifest.AppDockerImage{Image: "override"}},
			expected: manifest.Application{Docker: manifest.AppDockerImage{Image: "override"}},
		},
		"envs get merged": {
			base:     manifest.Application{Env: map[string]string{"base": "base"}},
			override: manifest.Application{Env: map[string]string{"override": "override"}},
			expected: manifest.Application{Env: map[string]string{"base": "base", "override": "override"}},
		},
		"override env priority": {
			base:     manifest.Application{Env: map[string]string{"base": "base"}},
			override: manifest.Application{Env: map[string]string{"override": "override", "base": "override"}},
			expected: manifest.Application{Env: map[string]string{"base": "override", "override": "override"}},
		},
		"buildpacks are strict override": {
			base:     manifest.Application{Buildpacks: []string{"java", "maven"}},
			override: manifest.Application{Buildpacks: []string{"node", "npm"}},
			expected: manifest.Application{Buildpacks: []string{"node", "npm"}},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			base := &tc.base
			err := base.Override(&tc.override)
			testutil.AssertNil(t, "override err", err)
			testutil.AssertEqual(t, "override", &tc.expected, base)
		})
	}
}

func ExampleApplication_Buildpack() {
	app := manifest.Application{}
	app.Buildpacks = []string{"java"}
	fmt.Println("One:", app.Buildpack())

	app.Buildpacks = []string{"maven", "java"}
	fmt.Println("Two:", app.Buildpack())

	// Output: One: java
	// Two: maven,java
}
