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

package describe

import (
	"os"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestJSONKeyToTitleCase(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"URL": {
			name:     "app.kubernetes.io/managed-by",
			expected: "app.kubernetes.io/managed-by",
		},
		"scaling key": {
			name:     "ephemeral-storage",
			expected: "ephemeral-storage",
		},
		"lastTransitionTime": {
			name:     "lastTransitionTime",
			expected: "Last Transition Time",
		},
		"numeric": {
			name:     "enableHTML5",
			expected: "Enable HTML5",
		},
		"resourceVersion": {
			name:     "resourceVersion",
			expected: "Resource Version",
		},
		"acronyms": {
			name:     "apiUrlUidOsbGuid",
			expected: "API URL UID OSB GUID",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := JSONKeyToTitleCase(tc.name)
			testutil.AssertEqual(t, "name", tc.expected, actual)
		})
	}
}

func ExampleUnstructured() {
	// example adapted from
	// https://github.com/kubernetes/kubectl/blob/master/pkg/describe/versioned/describe_test.go

	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion":           "v1",
			"kind":                 "Test",
			"sampleSample":         "present",
			"sample/sample":        "present",
			"sample-sample@sample": "present",
			"sample-sample":        "present",
			"sample1":              "present",
			"sample2":              "present",
			"metadata": map[string]interface{}{
				"name":              "MyName",
				"namespace":         "MyNamespace",
				"creationTimestamp": "2017-04-01T00:00:00Z",
				"resourceVersion":   123,
				"uid":               "00000000-0000-0000-0000-000000000001",
				"sample3":           "present",
			},
			"items": []interface{}{
				map[string]interface{}{
					"itemBool": true,
					"itemInt":  42,
				},
			},
			"url":    "http://localhost",
			"status": "ok",
		},
	}

	Unstructured(os.Stdout, resource)

	// Output: API Version:  v1
	// Items:
	//   Item Bool:  true
	//   Item Int:   42
	// Kind:  Test
	// Metadata:
	//   Creation Timestamp:  2017-04-01T00:00:00Z
	//   Name:                MyName
	//   Namespace:           MyNamespace
	//   Resource Version:    123
	//   Sample 3:            present
	//   UID:                 00000000-0000-0000-0000-000000000001
	// sample-sample:         present
	// sample-sample@sample:  present
	// sample/sample:         present
	// Sample 1:              present
	// Sample 2:              present
	// Sample Sample:         present
	// Status:                ok
	// URL:                   http://localhost
}

func ExampleUnstructuredStruct_nil() {
	UnstructuredStruct(os.Stdout, nil)

	// Output:
}

func ExampleUnstructuredStruct() {
	example := struct {
		IntField int            `json:"intField"`
		StrField string         `json:"strField"`
		ArrField []string       `json:"arrField"`
		MapField map[string]int `json:"strToInt"`
	}{
		IntField: 10,
		StrField: "some string",
		ArrField: []string{"one", "two", "three"},
		MapField: map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		},
	}

	UnstructuredStruct(os.Stdout, example)

	// Output: Arr Field:
	//   one
	//   two
	//   three
	// Int Field:  10
	// Str Field:  some string
	// Str To Int:
	//   A:  1
	//   B:  2
	//   C:  3
}
