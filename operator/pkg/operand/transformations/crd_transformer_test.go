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
	"kf-operator/pkg/operand/transformations"
	"testing"

	k8sTesting "kf-operator/pkg/testing/k8s"
	mfTesting "kf-operator/pkg/testing/manifestival"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/go-cmp/cmp"
)

func TestCRDTransformer(t *testing.T) {
	tests := []struct {
		name string
		in   mfTesting.Object
		want *unstructured.Unstructured
	}{{
		name: "not a crd",
		in:   k8sTesting.Deployment("test-deployment"),
		want: mfTesting.ToUnstructured(k8sTesting.Deployment("test-deployment")),
	}, {
		name: "a crd with no conversion strategy does not set preserveUnknownFields",
		in:   k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.NoneConverter)),
		want: mfTesting.ToUnstructured(k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.NoneConverter))),
	}, {
		name: "a crd with a webhook conversion strategy sets preserveUnknownFields",
		in:   k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.WebhookConverter)),
		want: mfTesting.ToUnstructured(k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.WebhookConverter)), withPreserveUnknownFields(false)),
	}, {
		name: "a crd with a webhook conversion strategy overwrites preserveUnknownFields",
		in:   k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.WebhookConverter), k8sTesting.WithPreserveUnknownFields(true)),
		want: mfTesting.ToUnstructured(k8sTesting.CRD("name", k8sTesting.WithConversionStrategy(apixv1.WebhookConverter)), withPreserveUnknownFields(false)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mfTesting.ToUnstructured(tt.in)

			transformer := transformations.CRDTransformer(context.Background())
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

func withPreserveUnknownFields(preserveUnknownFields bool) mfTesting.UnstructuredOption {
	return func(u *unstructured.Unstructured) {
		unstructured.SetNestedField(u.Object, false, "spec", "preserveUnknownFields")
	}
}
