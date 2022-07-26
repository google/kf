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
	certresources "knative.dev/pkg/webhook/certificates/resources"
)

func TestAPIServerSecretTransformer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     mftesting.Object
		want   *unstructured.Unstructured
		cert   string
		key    string
		cacert string
	}{
		{
			name: "not a secret",
			in:   k8s.ConfigMap("test-configmap"),
			want: mftesting.ToUnstructured(k8s.ConfigMap("test-configmap")),
		},
		{
			name: "not the upload-api-server-secret secret",
			in:   k8s.Secret("secret-something"),
			want: mftesting.ToUnstructured(k8s.Secret("secret-something")),
		},
		{
			name: "update secret with cert and key",
			in:   k8s.Secret("upload-api-server-secret"),
			want: mftesting.ToUnstructured(k8s.Secret("upload-api-server-secret", k8s.WithSecretData(map[string][]byte{
				certresources.ServerCert: []byte("some-cert"),
				certresources.ServerKey:  []byte("some-key"),
				certresources.CACert:     []byte("some-ca-cert"),
			}))),
			cert:   "some-cert",
			key:    "some-key",
			cacert: "some-ca-cert",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.in)

			transformer := APIServerSecretTransformer(
				context.Background(),
				[]byte(tt.cert),
				[]byte(tt.key),
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
