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

package kf

import (
	"context"
	"testing"

	"kf-operator/pkg/testing/k8s"
	mftesting "kf-operator/pkg/testing/manifestival"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAPIServerCertsTransformer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     mftesting.Object
		want   *unstructured.Unstructured
		cacert string
	}{
		{
			name: "not an APIService",
			in:   k8s.ConfigMap("test-configmap"),
			want: mftesting.ToUnstructured(k8s.ConfigMap("test-configmap")),
		},
		{
			name: "not the v1alpha1.upload.kf.dev APIService",
			in:   k8s.APIService("apiservice-something"),
			want: mftesting.ToUnstructured(k8s.APIService("apiservice-something")),
		},
		{
			name: "update APIService with CABundle",
			in:   k8s.APIService("v1alpha1.upload.kf.dev"),
			want: mftesting.ToUnstructured(
				k8s.APIService("v1alpha1.upload.kf.dev",
					k8s.WithAPIServiceInsecuretSkipTLSVerify(false),
					k8s.WithAPIServiceCABundle([]byte("some-ca-cert")),
				)),
			cacert: "some-ca-cert",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.in)

			transformer := APIServerCertsTransformer(
				context.Background(),
				[]byte(tt.cacert),
			)

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
