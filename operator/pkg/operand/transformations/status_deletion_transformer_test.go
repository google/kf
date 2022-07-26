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

package transformations

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	withStatus = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": "test",
			"spec":   "spec",
		},
	}
	withoutStatus = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": "spec",
		},
	}
)

func TestTransformStatusDeletion(t *testing.T) {
	transformer := TransformStatusDeletion()
	tests := []struct {
		name  string
		input *unstructured.Unstructured
		want  *unstructured.Unstructured
	}{{
		name:  "Delete status",
		input: withStatus,
		want:  withoutStatus,
	}, {
		name:  "No status",
		input: withoutStatus,
		want:  withoutStatus,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer(tt.input)
			if err != nil {
				t.Fatalf("Failed to transform: %v", err)
			}
			if diff := cmp.Diff(tt.input, tt.want); diff != "" {
				t.Fatalf("Unexpected diff: %s", diff)
			}
		})
	}
}
