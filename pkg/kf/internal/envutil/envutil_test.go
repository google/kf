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

package envutil_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/internal/envutil"
	"github.com/google/kf/pkg/kf/testutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVarsToMap(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		vars        []corev1.EnvVar
		expectedMap map[string]string
	}{
		"empty": {},
		"single": {
			vars: []corev1.EnvVar{
				{Name: "foo", Value: "bar"},
			},
			expectedMap: map[string]string{
				"foo": "bar",
			},
		},
		"overwrite": {
			vars: []corev1.EnvVar{
				{Name: "foo", Value: "bar"},
				{Name: "foo", Value: "bazz"},
			},
			expectedMap: map[string]string{
				"foo": "bazz",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "result", tc.expectedMap, envutil.EnvVarsToMap(tc.vars))
		})
	}
}

func TestMapToEnvVars(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		vars        map[string]string
		expectedEnv []corev1.EnvVar
	}{
		"empty": {},
		"single": {
			vars: map[string]string{
				"foo": "bar",
			},
			expectedEnv: []corev1.EnvVar{
				{Name: "foo", Value: "bar"},
			},
		},
		"multiple": {
			vars: map[string]string{
				"foo": "bazz",
				"bar": "boo",
			},
			expectedEnv: []corev1.EnvVar{
				{Name: "bar", Value: "boo"},
				{Name: "foo", Value: "bazz"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "result", tc.expectedEnv, envutil.MapToEnvVars(tc.vars))
		})
	}
}

func ExampleRemoveEnvVars() {
	envs := []corev1.EnvVar{
		{Name: "FOO", Value: "2"},
		{Name: "BAR", Value: "0"},
		{Name: "BAZZ", Value: "1"},
		{Name: "BAZZ", Value: "1.5"},
	}

	out := envutil.RemoveEnvVars([]string{"MISSING", "BAZZ"}, envs)
	for _, e := range out {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
	// Key FOO Value 2
}

func TestParseCLIEnvVars(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		vars        []string
		expectedEnv []corev1.EnvVar
		expectedErr error
	}{
		"empty": {},
		"single": {
			vars: []string{"foo=bar"},
			expectedEnv: []corev1.EnvVar{
				{Name: "foo", Value: "bar"},
			},
		},
		"overwrite": {
			vars: []string{"foo=bar", "foo=bazz"},
			expectedEnv: []corev1.EnvVar{
				{Name: "foo", Value: "bazz"},
			},
		},
		"too-many-equals": {
			vars:        []string{"foo=bar="},
			expectedErr: errors.New("malformed environment variable: foo=bar="),
		},
		"too-few-equals": {
			vars:        []string{"foo"},
			expectedErr: errors.New("malformed environment variable: foo"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			parsed, actualErr := envutil.ParseCLIEnvVars(tc.vars)
			if tc.expectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectedErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "result", tc.expectedEnv, parsed)
		})
	}

}

func ExampleDeduplicateEnvVars() {
	envs := []corev1.EnvVar{
		{Name: "FOO", Value: "2"},
		{Name: "BAR", Value: "0"},
		{Name: "BAZZ", Value: "1"},
		{Name: "BAZZ", Value: "1.5"},
	}

	out := envutil.DeduplicateEnvVars(envs)
	for _, e := range out {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
	// Key BAZZ Value 1.5
	// Key FOO Value 2
}

func ExampleNewJSONEnvVar() {
	env, err := envutil.NewJSONEnvVar("INVENTORY", map[string]bool{
		"Apples": true,
		"Bread":  false,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(env.Name, env.Value)

	// Output: INVENTORY {"Apples":true,"Bread":false}
}

func ExampleGetServiceEnvVars() {
	var service serving.Service
	envutil.SetServiceEnvVars(&service, []corev1.EnvVar{
		{Name: "FOO", Value: "2"},
		{Name: "BAR", Value: "0"},
	})

	env := envutil.GetServiceEnvVars(&service)

	for _, e := range env {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key FOO Value 2
	// Key BAR Value 0
}

func ExampleGetServiceEnvVars_emptyService() {
	env := envutil.GetServiceEnvVars(nil)

	fmt.Println(env)

	// Output: []
}
