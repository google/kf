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

package commandtools

import (
	"fmt"
	"strings"
	"testing"
)

func newFakeExecuteTerminalCommandFunc(koOutput string, gcloudOutput string) executeTerminalCommandFunc {
	return func(command string, arguments ...string) string {
		if command == "ko" {
			return koOutput
		} else if command == "gcloud" {
			return gcloudOutput
		} else {
			panic(fmt.Sprintf("Unknown command %q with arguments %#v", command, arguments))
		}
	}
}

func newFakeGlobFunc(filePaths []string, globError error) globFunc {
	return func(pattern string) ([]string, error) {
		return filePaths, globError
	}
}

func newFakeCountLinesFunc(fileLinesCount map[string]int) countLinesFunc {
	return func(filePath string) int {
		linesCount, ok := fileLinesCount[filePath]
		if !ok {
			panic(fmt.Sprintf("Cannot find file in path %q", filePath))
		}

		return linesCount
	}
}

func newFakeWriteFileFunc(expectedContent string, t *testing.T) writeLinesFunc {
	return func(filePath, content string) {
		if content != expectedContent {
			t.Fatalf("Called writeFile with %q, but expected %q", content, expectedContent)
		}
	}
}

func TestKoResolver(t *testing.T) {
	tests := []struct {
		description string

		configPath     string
		tag            string
		resultFilePath string
		version        string
		substitutions  []Substitution

		fileLinesCount      map[string]int
		globError           error
		koCommandOutput     string
		gcloudCommandOutput string

		expectedContainedPanicMessage string
		expectedFilteredResult        string
		expectedResult                string
	}{
		{
			description:                   "Glob failed to fetch yaml files",
			configPath:                    "/config",
			globError:                     fmt.Errorf("glob error: Permission denied"),
			expectedContainedPanicMessage: "glob error",
		},
		{
			description:                   "No yaml files matched",
			configPath:                    "/config",
			fileLinesCount:                map[string]int{},
			expectedContainedPanicMessage: "no yaml files",
		},
		{
			description: "File produced by ko is empty",
			configPath:  "/config",
			fileLinesCount: map[string]int{
				"/config/namespace.yaml": 10,
			},
			koCommandOutput:               "",
			expectedContainedPanicMessage: "empty",
		},
		{
			description: "Number of lines produced by ko is less than expected min",
			configPath:  "/config",
			fileLinesCount: map[string]int{
				"/config/namespace.yaml": 10,
			},
			koCommandOutput: "apiVersion: v1\n" +
				"kind: Namespace\n" +
				"metadata:\n" +
				"name: test-namespace",
			expectedContainedPanicMessage: "less than expected min",
		},
		{
			description: "Number of lines produced by ko is greater than expected max",
			configPath:  "/config",
			fileLinesCount: map[string]int{
				"/config/namespace.yaml": 1,
			},
			koCommandOutput: "apiVersion: v1\n" +
				"kind: Namespace\n" +
				"metadata:\n" +
				"name: test-namespace",
			expectedContainedPanicMessage: "greater than expected max",
		},
		{
			description:    "koResolver resolve files successfully",
			configPath:     "/config",
			tag:            "0.0.0-gke.0",
			resultFilePath: "/result",
			version:        "v0.0.0-gke.0",
			fileLinesCount: map[string]int{
				"/config/namespace.yaml":  6,
				"/config/deployment.yaml": 13,
			},
			substitutions: []Substitution{
				{
					Origin: "name: test",
					Target: "name: test-namespace",
				},
			},
			koCommandOutput: `apiVersion: v1
kind: Namespace
metadata:
  labels:
    operator.knative.dev/release: devel
    name: test
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    operator.knative.dev/release: devel
spec:
  template:
    metadata:
      labels:
        operator.knative.dev/release: devel
  spec:
    containers:
      image: gke.gcr.io/test-image:0.0.0-gke.0
---`,
			expectedFilteredResult: `apiVersion: v1
kind: Namespace
metadata:
  labels:
    operator.knative.dev/release: "0.0.0-gke.0"
    name: test-namespace
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    operator.knative.dev/release: "0.0.0-gke.0"
spec:
  template:
    metadata:
      labels:
        operator.knative.dev/release: "0.0.0-gke.0"
  spec:
    containers:
      image: gke.gcr.io/test-image:0.0.0-gke.0
---
`,
			expectedResult: `apiVersion: v1
kind: Namespace
metadata:
  labels:
    operator.knative.dev/release: "0.0.0-gke.0"
    name: test
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    operator.knative.dev/release: "0.0.0-gke.0"
spec:
  template:
    metadata:
      labels:
        operator.knative.dev/release: "0.0.0-gke.0"
  spec:
    containers:
      image: gke.gcr.io/test-image:0.0.0-gke.0
---`,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.expectedContainedPanicMessage != "" {
				defer func() {
					r := recover()
					if r == nil {
						t.Fatalf("Not paniced, but expected to panic with %q contained", test.expectedContainedPanicMessage)
					}
					err, ok := r.(string)
					if !ok {
						t.Fatalf("Paniced with type %T, but expected string type", r)
					}
					if !strings.Contains(err, test.expectedContainedPanicMessage) {
						t.Fatalf("Got panic %q, but expeced to contain %q", err, test.expectedContainedPanicMessage)
					}
				}()
			}

			// Get file paths from files.
			filePaths := make([]string, 0, len(test.fileLinesCount))
			for filePath := range test.fileLinesCount {
				filePaths = append(filePaths, filePath)
			}

			koResolver := &KoResolver{
				executeTerminalCommand: newFakeExecuteTerminalCommandFunc(test.koCommandOutput, test.gcloudCommandOutput),
				glob:                   newFakeGlobFunc(filePaths, test.globError),
				countLines:             newFakeCountLinesFunc(test.fileLinesCount),
				writeFile:              newFakeWriteFileFunc(test.expectedFilteredResult, t),
			}

			result := koResolver.Resolve(test.configPath, test.tag, test.resultFilePath, test.version, test.substitutions...)
			if result != test.expectedResult {
				t.Fatalf("Got result %q, but expect %q", result, test.expectedResult)
			}
		})
	}
}
