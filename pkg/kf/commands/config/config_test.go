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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	homedir "github.com/mitchellh/go-homedir"
)

func TestParamsPath(t *testing.T) {
	userHome, _ := homedir.Dir()

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
