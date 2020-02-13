// Copyright 2019 Google LLC
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
// limitations under the License.

package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/homedir"
)

func TestParamsPath(t *testing.T) {
	userHome := homedir.HomeDir()

	cases := map[string]struct {
		path     string
		expected string
	}{
		"override": {
			path:     "some/custom/path.yaml",
			expected: "some/custom/path.yaml",
		},
		"default": {
			path:     "",
			expected: path.Join(userHome, ".kf"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := paramsPath(tc.path)
			testutil.AssertEqual(t, "paths", tc.expected, actual)
		})
	}
}

func ExampleWrite() {
	dir, err := ioutil.TempDir("", "kfcfg")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	configFile := path.Join(dir, "kf.yaml")

	{
		toWrite := &KfParams{
			Namespace: "my-namespace",
		}

		if err := Write(configFile, toWrite); err != nil {
			panic(err)
		}
	}

	{
		toRead, err := NewKfParamsFromFile(configFile)
		if err != nil {
			panic(err)
		}

		fmt.Println("Read namespace:", toRead.Namespace)
	}

	// Output: Read namespace: my-namespace
}

func TestNewDefaultKfParams(t *testing.T) {
	cases := map[string]struct {
		configPathEnv string
		expected      KfParams
	}{
		"KUBECONFIG set": {
			configPathEnv: "kube-config.yml",
			expected: KfParams{
				KubeCfgFile: "kube-config.yml",
			},
		},
		"KUBECONFIG not set": {
			configPathEnv: "",
			expected: KfParams{
				KubeCfgFile: filepath.Join(homedir.HomeDir(), ".kube", "config"),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if len(tc.configPathEnv) != 0 {
				os.Setenv("KUBECONFIG", tc.configPathEnv)
			} else {
				os.Unsetenv("KUBECONFIG")
			}
			defaultConfig := NewDefaultKfParams()

			testutil.AssertEqual(t, "config", &tc.expected, defaultConfig)
		})
	}

}

func TestLoad(t *testing.T) {

	defaultConfig := *NewDefaultKfParams()

	cases := map[string]struct {
		configFile  string
		override    KfParams
		expected    KfParams
		expectedErr error
	}{
		"empty config": {
			configFile: "empty-config.yml",
			expected:   defaultConfig,
		},
		"missing config": {
			configFile:  "missing-config.yml",
			expectedErr: errors.New("open testdata/missing-config.yml: no such file or directory"),
		},
		"overrides": {
			configFile: "custom-config.yml",
			override: KfParams{
				Namespace:   "foo",
				KubeCfgFile: "kubecfg",
			},
			expected: KfParams{
				Namespace:   "foo",
				KubeCfgFile: "kubecfg",
			},
		},
		"populated config": {
			configFile: "custom-config.yml",
			expected: KfParams{
				Namespace:   "customns",
				KubeCfgFile: "customkubecfg",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := Load(path.Join("testdata", tc.configFile), &tc.override)
			if tc.expectedErr != nil || err != nil {
				testutil.AssertErrorsEqual(t, tc.expectedErr, err)
				return
			}

			testutil.AssertEqual(t, "config", &tc.expected, actual)
		})
	}

}

func ExampleKfParams_GetTargetSpaceOrDefault() {
	target := &v1alpha1.Space{}
	target.Name = "cached-target"

	p := &KfParams{
		TargetSpace: target,
	}

	space, err := p.GetTargetSpaceOrDefault()
	fmt.Println("Space:", space.Name)
	fmt.Println("Error:", err)

	// Output: Space: cached-target
	// Error: <nil>
}

func ExampleKfParams_SetTargetSpaceToDefault() {
	defaultSpace := &v1alpha1.Space{}
	defaultSpace.SetDefaults(context.Background())

	p := &KfParams{}
	p.SetTargetSpaceToDefault()

	fmt.Printf("Set to default: %v\n", reflect.DeepEqual(p.TargetSpace, defaultSpace))

	// Output: Set to default: true
}

func TestKfParams_cacheSpace(t *testing.T) {
	goodSpace := &v1alpha1.Space{}
	goodSpace.Name = "test-space"

	defaultSpace := &v1alpha1.Space{}
	defaultSpace.SetDefaults(context.Background())

	cases := map[string]struct {
		space *v1alpha1.Space
		err   error

		expectSpace       *v1alpha1.Space
		expectErr         error
		expectUpdateSpace bool
	}{
		"no error": {
			space:       goodSpace,
			expectSpace: goodSpace,
		},
		"not found error": {
			err:         apierrs.NewNotFound(v1alpha1.Resource("sources"), ""),
			expectSpace: defaultSpace,
		},
		"other error": {
			err:       errors.New("api connection error"),
			expectErr: errors.New("couldn't get the Space \"test-space\": api connection error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			p := &KfParams{}
			p.Namespace = "test-space"

			actualSpace, actualErr := p.cacheSpace(tc.space, tc.err)

			testutil.AssertEqual(t, "spaces", tc.expectSpace, actualSpace)
			testutil.AssertEqual(t, "errors", tc.expectErr, actualErr)
			testutil.AssertEqual(t, "TargetSpace", tc.expectSpace, p.TargetSpace)
		})
	}
}
