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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/ptr"
)

func TestNewFromReader(t *testing.T) {
	relativePathRoot := filepath.Dir(filepath.FromSlash("some/manifest/dir/manifest.yaml"))

	cases := map[string]struct {
		fileContent string
		expected    *manifest.Manifest
		expectErr   error
	}{
		"app name": {
			fileContent: `---
applications:
- name: MY-APP
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

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
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name:       "MY-APP",
						Stack:      "cflinuxfs3",
						Buildpacks: []string{"java", "node"},
					},
				},
			},
		},
		"routes": {
			fileContent: `---
applications:
- name: MY-APP
  routes:
  - route: some-route.domain.com
  - route: "*.another-route.domain.com"
  - route: app-port.domain.com
    appPort: 9999
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name: "MY-APP",
						Routes: []manifest.Route{
							{
								Route: "some-route.domain.com",
							},
							{
								Route: "*.another-route.domain.com",
							},
							{
								Route:   "app-port.domain.com",
								AppPort: 9999,
							},
						},
					},
				},
			},
		},
		"legacy-buildpack": {
			fileContent: `---
applications:
- name: MY-APP
  stack: cflinuxfs3
  buildpack: java
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name:            "MY-APP",
						Stack:           "cflinuxfs3",
						LegacyBuildpack: "java",
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
				RelativePathRoot: relativePathRoot,

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
		"cf style command": {
			fileContent: `---
applications:
- name: CUSTOM_START
  command: rake run $VAR
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name:    "CUSTOM_START",
						Command: "rake run $VAR",
					},
				},
			},
		},
		"docker style command": {
			fileContent: `---
applications:
- name: CUSTOM_START
  entrypoint: python
  args: ["-m", "SimpleHTTPServer"]
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name: "CUSTOM_START",
						KfApplicationExtension: manifest.KfApplicationExtension{
							Entrypoint: "python",
							Args:       []string{"-m", "SimpleHTTPServer"},
						},
					},
				},
			},
		},
		"dockerfile": {
			fileContent: `---
applications:
- name: MY-APP
  dockerfile:
    path: "foo/Dockerfile"
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,

				Applications: []manifest.Application{
					{
						Name: "MY-APP",
						KfApplicationExtension: manifest.KfApplicationExtension{
							Dockerfile: manifest.Dockerfile{
								Path: "foo/Dockerfile",
							},
						},
					},
				},
			},
		},
		"malformed": {
			fileContent: `---
applications:
  dockerfile:
    path: "foo/Dockerfile"
`,
			expected:  nil,
			expectErr: errors.New("manifest appears invalid: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal object into Go struct field Manifest.applications of type []manifest.Application"),
		},
		"extra properties parse correctly": {
			fileContent: `---
applications:
- name: MY-APP
  extraproperty: 42
`,
			expected: &manifest.Manifest{
				RelativePathRoot: relativePathRoot,
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
			actual, err := manifest.NewFromReader(context.Background(), strings.NewReader(tc.fileContent), relativePathRoot, nil)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)
			testutil.AssertEqual(t, "manifest", tc.expected, actual)
		})
	}
}

func TestCheckForManifest(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		path     string
		expected *manifest.Manifest
	}{
		"yml": {
			path: "testdata/manifest-dir",
			expected: &manifest.Manifest{
				RelativePathRoot: filepath.Join(pwd, filepath.FromSlash("testdata/manifest-dir")),
				Applications: []manifest.Application{
					{
						Name: "MY-APP",
					},
				},
			},
		},
		"yaml": {
			path: "testdata/manifest-dir-2",
			expected: &manifest.Manifest{
				RelativePathRoot: filepath.Join(pwd, filepath.FromSlash("testdata/manifest-dir-2")),
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
			cwd, err := os.Getwd()
			testutil.AssertNil(t, "error", err)
			testutil.AssertNil(t, "chdir", os.Chdir(tc.path))
			actual, err := manifest.CheckForManifest(context.Background(), nil)
			testutil.AssertNil(t, "chdir", os.Chdir(cwd))
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
		"no start, no override": {
			base:     manifest.Application{KfApplicationExtension: manifest.KfApplicationExtension{NoStart: ptr.Bool(true)}},
			override: manifest.Application{},
			expected: manifest.Application{KfApplicationExtension: manifest.KfApplicationExtension{NoStart: ptr.Bool(true)}},
		},
		"no start, override": {
			base:     manifest.Application{KfApplicationExtension: manifest.KfApplicationExtension{NoStart: ptr.Bool(false)}},
			override: manifest.Application{KfApplicationExtension: manifest.KfApplicationExtension{NoStart: ptr.Bool(true)}},
			expected: manifest.Application{KfApplicationExtension: manifest.KfApplicationExtension{NoStart: ptr.Bool(true)}},
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
	app.LegacyBuildpack = "hidden-legacy-buildpack"
	fmt.Println("Legacy:", app.Buildpack())

	app.Buildpacks = []string{"java"}
	fmt.Println("One:", app.Buildpack())

	app.Buildpacks = []string{"maven", "java"}
	fmt.Println("Two:", app.Buildpack())

	// Output: Legacy: hidden-legacy-buildpack
	// One: java
	// Two: maven,java
}

func ExampleApplication_BuildpacksSlice() {
	app := manifest.Application{}
	app.LegacyBuildpack = "hidden-legacy-buildpack"
	fmt.Println("Legacy:", app.BuildpacksSlice())

	app.Buildpacks = []string{"java"}
	fmt.Println("One:", app.BuildpacksSlice())

	app.Buildpacks = []string{"maven", "java"}
	fmt.Println("Two:", app.BuildpacksSlice())

	// Output: Legacy: [hidden-legacy-buildpack]
	// One: [java]
	// Two: [maven java]
}

func ExampleApplication_CommandArgs() {
	app := manifest.Application{}
	fmt.Printf("Blank: %v\n", app.CommandArgs())

	app = manifest.Application{
		Command: "start.sh && exit 1",
	}
	fmt.Printf("Command: %v\n", app.CommandArgs())

	app = manifest.Application{
		KfApplicationExtension: manifest.KfApplicationExtension{
			Args: []string{"-m", "SimpleHTTPServer"},
		},
	}
	fmt.Printf("Args: %v\n", app.CommandArgs())

	// Output: Blank: []
	// Command: [start.sh && exit 1]
	// Args: [-m SimpleHTTPServer]
}

func ExampleApplication_CommandEntrypoint() {
	app := manifest.Application{}
	fmt.Printf("Blank: %v\n", app.CommandEntrypoint())

	app = manifest.Application{
		KfApplicationExtension: manifest.KfApplicationExtension{
			Entrypoint: "python",
		},
	}
	fmt.Printf("Entrypoint: %v\n", app.CommandEntrypoint())

	// Output: Blank: []
	// Entrypoint: [python]
}

func ExampleApplication_WriteWarnings() {
	app := manifest.Application{
		Name: "hello_world",
		KfApplicationExtension: manifest.KfApplicationExtension{
			NoStart: ptr.Bool(false),
			Ports: manifest.AppPortList{
				{Port: 8080, Protocol: "tcp"},
			},
		},
	}

	app.WriteWarnings(configlogging.SetupLogger(context.Background(), os.Stdout))

	// Output:
	// WARN The field(s) [no-start ports] are Kf-specific manifest extensions and may change.
	// WARN Kf supports TCP ports but currently only HTTP Routes. TCP ports can be reached on the App's cluster-internal app-<name>.<space>.svc.cluster.local address.
	// WARN Underscores ('_') in names are not allowed in Kubernetes. Replacing with hyphens ('-')...
}

func TestApp(t *testing.T) {
	cases := map[string]struct {
		applications  []manifest.Application
		appName       string
		expected      *manifest.Application
		expectedError bool
	}{
		"happy path; single App": {
			applications: []manifest.Application{
				{Name: "random"},
			},
			appName:       "abc",
			expected:      &manifest.Application{Name: "abc"},
			expectedError: false,
		},
		"empty name; single App": {
			applications: []manifest.Application{
				{Name: "random"},
			},
			appName:       "",
			expected:      &manifest.Application{Name: "random"},
			expectedError: false,
		},
		"Multiple Apps": {
			applications: []manifest.Application{
				{Name: "random"},
				{Name: "random1"},
			},
			appName:       "",
			expected:      nil,
			expectedError: true,
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			m := manifest.Manifest{
				Applications: tc.applications,
			}
			actual, err := m.App(tc.appName)
			testutil.AssertEqual(t, "App", tc.expected, actual)
			if tc.expectedError {
				testutil.AssertNotNil(t, "App", err)
			}
		})
	}
}
