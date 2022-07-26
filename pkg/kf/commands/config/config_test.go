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
	"strings"
	"testing"

	apiconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	kubeclient "knative.dev/pkg/client/injection/kube/client/fake"
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
			Space: "my-namespace",
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

		fmt.Println("Read namespace:", toRead.Space)
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
				Space:       "foo",
				KubeCfgFile: "kubecfg",
			},
			expected: KfParams{
				Space:       "foo",
				KubeCfgFile: "kubecfg",
			},
		},
		"populated config": {
			configFile: "custom-config.yml",
			expected: KfParams{
				Space:       "customns",
				KubeCfgFile: "", // this changed after Kf v2.1.0 so Kf no long caches Kubeconfig.
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

func ExampleKfParams_GetTargetSpace() {
	target := &v1alpha1.Space{}
	target.Name = "cached-target"

	p := &KfParams{
		TargetSpace: target,
	}

	space, err := p.GetTargetSpace(context.Background())
	fmt.Println("Space:", space.Name)
	fmt.Println("Error:", err)

	// Output: Space: cached-target
	// Error: <nil>
}

func TestKfParams_cacheSpace(t *testing.T) {
	goodSpace := &v1alpha1.Space{}
	goodSpace.Name = "test-space"

	defaultSpace := &v1alpha1.Space{}
	defaultSpace.SetDefaults(apiconfig.DefaultConfigContext(context.Background()))

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
			err:       apierrs.NewNotFound(v1alpha1.Resource("builds"), ""),
			expectErr: errors.New(`Space "test-space" doesn't exist`),
		},
		"other error": {
			err:       errors.New("api connection error"),
			expectErr: errors.New("couldn't get the Space \"test-space\": api connection error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			p := &KfParams{}
			p.Space = "test-space"

			actualSpace, actualErr := p.cacheSpace(tc.space, tc.err)
			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "spaces", tc.expectSpace, actualSpace)
			testutil.AssertEqual(t, "TargetSpace", tc.expectSpace, p.TargetSpace)
		})
	}
}

func TestFeatureFlagWarning(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		namespace   *v1.Namespace
		expectedOut string
	}{
		"nominal": {
			namespace:   namespace(v1alpha1.KfNamespace, ""),
			expectedOut: "",
		},
		"missing flags": {
			namespace: (func() *v1.Namespace {
				ns := namespace(v1alpha1.KfNamespace, "")
				delete(ns.Annotations, v1alpha1.FeatureFlagsAnnotation)
				return ns
			}()),
			expectedOut: fmt.Sprintf("WARN Unable to read feature flags from server.\n"),
		},
		"bad flags": {
			namespace: (func() *v1.Namespace {
				ns := namespace(v1alpha1.KfNamespace, "")
				ns.Annotations[v1alpha1.FeatureFlagsAnnotation] = ""
				return ns
			}()),
			expectedOut: "WARN Invalid feature flag config: unexpected end of JSON input\n",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := injection.WithInjection(context.Background(), t)
			client := kubeclient.Get(ctx)
			_, err := client.
				CoreV1().
				Namespaces().
				Create(ctx, tc.namespace, metav1.CreateOptions{})
			testutil.AssertErrorsEqual(t, nil, err)

			buf := &strings.Builder{}
			ctx = configlogging.SetupLogger(ctx, buf)
			var p KfParams
			p.FeatureFlags(ctx)

			actual := buf.String()
			testutil.AssertEqual(t, "warning messages", tc.expectedOut, actual)
		})
	}
}

func namespace(namespace, status string) *v1.Namespace {
	return &v1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: namespace,
		Labels: map[string]string{
			v1alpha1.NameLabel: namespace,
		},
		Annotations: map[string]string{
			v1alpha1.FeatureFlagsAnnotation: "{}",
		},
	}}
}
