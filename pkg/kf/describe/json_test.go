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

	"github.com/google/kf/pkg/kf/testutil"
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
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := JSONKeyToTitleCase(tc.name)
			testutil.AssertEqual(t, "name", tc.expected, actual)
		})
	}
}

func ExampleJSONData() {
	JSONData(os.Stdout, []byte(`{
    "apiVersion": "kf.dev/v1alpha1",
    "kind": "App",
    "metadata": {
        "creationTimestamp": "2019-09-19T22:49:22Z",
        "generation": 1,
        "labels": {
            "app.kubernetes.io/managed-by": "kf"
        },
        "name": "head",
        "namespace": "joseph",
        "resourceVersion": "65838132",
        "selfLink": "/apis/kf.dev/v1alpha1/namespaces/joseph/apps/head",
        "uid": "b8d67ca3-db2f-11e9-b478-42010a8000dd"
    }
}`))

	// Output: foo
}

func ExampleJSONData_array() {
	JSONData(os.Stdout, []byte(`{"conditions": [
            {
                "lastTransitionTime": "2019-09-19T22:49:52Z",
                "status": "True",
                "type": "EnvVarSecretReady"
            },
            {
                "lastTransitionTime": "2019-09-23T16:59:09Z",
                "status": "True",
                "type": "KnativeServiceReady"
            },
            {
                "lastTransitionTime": "2019-09-23T16:59:09Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastTransitionTime": "2019-09-19T22:49:52Z",
                "severity": "Info",
                "status": "True",
                "type": "ServiceBindingsReady"
            },
            {
                "lastTransitionTime": "2019-09-19T22:49:52Z",
                "status": "True",
                "type": "SourceReady"
            },
            {
                "lastTransitionTime": "2019-09-19T22:49:22Z",
                "status": "True",
                "type": "SpaceReady"
            }
        ]}`))

	// Output: foo
}
