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

	"kf-operator/pkg/apis/kfsystem/kf"
	"kf-operator/pkg/testing/k8s"
	mfTesting "kf-operator/pkg/testing/manifestival"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/go-cmp/cmp"
)

func TestAddFeatureFlags(t *testing.T) {
	tests := []struct {
		name         string
		in           mfTesting.Object
		featureFlags kf.FeatureFlagToggles
		want         *unstructured.Unstructured
	}{{
		name: "not desired namespace",
		in:   k8s.Namespace("test-namespace"),
		want: mfTesting.ToUnstructured(k8s.Namespace("test-namespace")),
	}, {
		name: "not desired kind",
		in:   k8s.Deployment("test-deployment"),
		want: mfTesting.ToUnstructured(k8s.Deployment("test-deployment")),
	}, {
		name: "no feature flags",
		in:   k8s.Namespace("kf"),
		want: mfTesting.ToUnstructured(k8s.Namespace("kf")),
	}, {
		name: "desired kind and name",
		in:   k8s.Namespace("kf"),
		featureFlags: kf.FeatureFlagToggles{
			"testfeature": true,
		},
		want: mfTesting.ToUnstructured(
			k8s.Namespace(
				"kf",
				k8s.WithNamespaceAnnotation(map[string]string{FeatureFlagsAnnotation: `{"testfeature":true}`}),
			)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mfTesting.ToUnstructured(tt.in)

			transformer := AddFeatureFlags(
				context.Background(),
				tt.featureFlags)

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
