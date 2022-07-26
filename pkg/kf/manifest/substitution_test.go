// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestApplySubstitution(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input     string
		variables map[string]interface{}

		wantErr error
	}{
		"missing substitution": {
			input:     "name: ((name))",
			variables: nil,
			wantErr:   errors.New("no variable found for key: ((name))"),
		},
		"nested map": {
			input: `
env:
  ((key)): ((value))`,
			variables: map[string]interface{}{"key": 1, "value": 2},
		},
		"nested int map": {
			input: `
env:
  1: ((value))
  2: ((value))`,
			variables: map[string]interface{}{"key": 1, "value": 2},
		},
		"whole thing": {
			input: `
applications:
- name: ((name))
  env:
    ((db type))_CXN: 127.0.0.1:((port))
  instances: ((instances))
  services: ((services))`,
			variables: map[string]interface{}{
				"name":      "my-app",
				"port":      8080,
				"db type":   "MYSQL",
				"instances": 3,
				"services":  []string{"a", "b", "c"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotSub, gotErr := ApplySubstitution([]byte(tc.input), tc.variables)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			testutil.AssertGoldenContext(t, "yaml", gotSub, map[string]interface{}{
				"input":     tc.input,
				"variables": tc.variables,
			})
		})
	}
}

func Test_substitute(t *testing.T) {
	t.Parallel()

	//func substitute(obj interface{}, variables map[string]interface{}) (interface{}, error) {
	cases := map[string]struct {
		obj       interface{}
		variables map[string]interface{}

		wantObj interface{}
		wantErr error
	}{
		"nil object": {
			obj: nil,
			variables: map[string]interface{}{
				"foo": "bar",
			},
			wantObj: nil,
			wantErr: nil,
		},
		"string substitution": {
			obj: "((foo))-((foo))",
			variables: map[string]interface{}{
				"foo": "bar",
			},
			wantObj: "bar-bar",
			wantErr: nil,
		},
		"string substitution middle fails": {
			obj: "((foo))((bar))((foo))",
			variables: map[string]interface{}{
				"foo": "bar",
			},
			wantErr: errors.New("no variable found for key: ((bar))"),
		},
		"full substitution replaces type": {
			obj: "((ports))",
			variables: map[string]interface{}{
				"ports": []int{8080, 9090},
			},
			wantObj: []int{8080, 9090},
			wantErr: nil,
		},
		"array replacement": {
			obj: []interface{}{"((first))", "((second))"},
			variables: map[string]interface{}{
				"first":  1,
				"second": 2,
			},
			wantObj: []interface{}{1, 2},
			wantErr: nil,
		},
		"map replacement": {
			obj: map[string]interface{}{"key-((first))": "value-((second))"},
			variables: map[string]interface{}{
				"first":  1,
				"second": 2,
			},
			wantObj: map[string]interface{}{"key-1": "value-2"},
			wantErr: nil,
		},
		"bad replacement": {
			obj:       "key-((first))",
			variables: map[string]interface{}{},
			wantObj:   nil,
			wantErr:   errors.New("no variable found for key: ((first))"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotObj, gotErr := substitute(tc.obj, tc.variables)

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertEqual(t, "object", tc.wantObj, gotObj)
		})
	}
}

func Test_findSubstitution(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		key       string
		variables map[string]interface{}

		wantSub interface{}
		wantErr error
	}{
		"found": {
			key: "((key))",
			variables: map[string]interface{}{
				"key": 42,
			},
			wantSub: 42,
		},
		"not found": {
			key: "((dne))",
			variables: map[string]interface{}{
				"key": 42,
			},
			wantSub: nil,
			wantErr: errors.New("no variable found for key: ((dne))"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotSub, gotErr := findSubstitution(tc.key, tc.variables)

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertEqual(t, "variable", tc.wantSub, gotSub)
		})
	}
}
