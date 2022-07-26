// Copyright 2020 Google LLC
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

package secrets

import (
	"encoding/json"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

func TestBuildParamsSecret(t *testing.T) {
	tests := map[string]struct {
		owner  kmeta.OwnerRefable
		name   string
		params json.RawMessage

		want *v1.Secret
	}{
		"nominal": {
			owner: &v1alpha1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db",
					Namespace: "some-ns",
				},
			},
			name:   "serviceinstance-test-db-params-2paukuefdpllj",
			params: json.RawMessage("{}"),
			want: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "serviceinstance-test-db-params-2paukuefdpllj",
					Namespace: "some-ns",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "kf.dev/v1alpha1",
							Kind:               "ServiceInstance",
							Name:               "test-db",
							Controller:         ptr.Bool(true),
							BlockOwnerDeletion: ptr.Bool(true),
						},
					},
				},
				Data: map[string][]byte{
					"params": json.RawMessage("{}"),
				},
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			got := BuildParamsSecret(tc.owner, tc.name, tc.params)
			testutil.AssertEqual(t, "secret", tc.want, got)
		})
	}
}

func TestCreateJSONPatch(t *testing.T) {
	tests := map[string]struct {
		params json.RawMessage
		want   []byte
	}{
		"nominal": {
			params: json.RawMessage("{}"),
			want:   []byte(`[{"op":"replace","path":"/data/params","value":"e30="}]`),
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			got, _ := CreateJSONPatch(tc.params)
			testutil.AssertEqual(t, "jsonPatch", tc.want, got)
		})
	}
}
