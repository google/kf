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

package transformations_test

import (
	"context"
	"testing"

	"kf-operator/pkg/operand/transformations"
	"kf-operator/pkg/testing/k8s"
	mfTesting "kf-operator/pkg/testing/manifestival"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/go-cmp/cmp"
)

func TestAddAnnotation(t *testing.T) {
	tests := []struct {
		name       string
		in         mfTesting.Object
		kind       string
		objectName string
		want       *unstructured.Unstructured
	}{{
		name:       "not desired kind",
		in:         k8s.Deployment("test-deployment"),
		kind:       "SomethingElse",
		objectName: "something",
		want:       mfTesting.ToUnstructured(k8s.Deployment("test-deployment")),
	}, {
		name:       "not desired name",
		in:         k8s.Deployment("test-deployment"),
		kind:       "SomethingElse",
		objectName: "something",
		want:       mfTesting.ToUnstructured(k8s.Deployment("test-deployment")),
	}, {
		name:       "desired kind and name",
		in:         k8s.Deployment("test-deployment"),
		kind:       "Deployment",
		objectName: "test-deployment",
		want: mfTesting.ToUnstructured(
			k8s.Deployment(
				"test-deployment",
				k8s.WithDeploymentAnnotation(map[string]string{"test": "test"}),
			)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mfTesting.ToUnstructured(tt.in)

			transformer := transformations.AddAnnotation(
				context.Background(),
				tt.kind,
				tt.objectName,
				map[string]string{
					"test": "test",
				})

			err := transformer(u)
			if err != nil {
				t.Error("Got error", err)
			}

			if diff := cmp.Diff(tt.want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}
