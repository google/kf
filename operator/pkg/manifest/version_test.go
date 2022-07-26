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

package manifest

import (
	"testing"

	mf "github.com/manifestival/manifestival"

	. "kf-operator/pkg/testing/k8s"
	. "kf-operator/pkg/testing/manifestival"

	"github.com/google/go-cmp/cmp"
)

const testLabel = "test-label"

func TestVersionFromLabel(t *testing.T) {
	tests := []struct {
		name  string
		input mf.Slice
		want  string
	}{{
		name: "returns default with no resources",
		want: "test-default",
	}, {
		name: "returns default when resources do not have label",
		input: mf.Slice{
			*ToUnstructured(AddLabel(Namespace("test"), "other-label", "other-value")),
		},
		want: "test-default",
	}, {
		name: "returns the label when resources have label",
		input: mf.Slice{
			*ToUnstructured(AddLabel(Namespace("test"), "test-label", "explicit-value")),
		},
		want: "explicit-value",
	}, {
		name: "returns the first label when resources have label",
		input: mf.Slice{
			*ToUnstructured(AddLabel(Namespace("test"), "test-label", "first-explicit-value")),
			*ToUnstructured(AddLabel(Namespace("test"), "test-label", "second-explicit-value")),
		},
		want: "first-explicit-value",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest, err := mf.ManifestFrom(test.input)

			if err != nil {
				t.Errorf("mf.ManifestFrom error: %v", err)
			}

			got := VersionFromLabel(&manifest, testLabel, "test-default")
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("VersionFromLabel (-want, +got) = %v", diff)
			}
		})
	}
}

func TestVersionFromDeploymentImage(t *testing.T) {
	tests := []struct {
		name  string
		input mf.Slice
		want  string
	}{{
		name: "returns default with no resources",
		want: "test-default",
	}, {
		name: "returns default when resources do not have an image",
		input: mf.Slice{
			*ToUnstructured(Deployment("test")),
		},
		want: "test-default",
	}, {
		name: "returns default when resources image does not have a tag",
		input: mf.Slice{
			*ToUnstructured(Deployment("test", WithDeploymentContainer(Container("test-name", "test-image")))),
		},
		want: "test-default",
	}, {
		name: "returns default when resources image uses a digest",
		input: mf.Slice{
			*ToUnstructured(Deployment("test", WithDeploymentContainer(Container("test-name", "gcr.io/test/test-image@sha256:fb41ffa50dc4924d719e4b7e1d889b1d42444ac69bd4b3d6ee6371495ad431ce")))),
		},
		want: "test-default",
	}, {
		name: "returns the image tag when the image does have a tag",
		input: mf.Slice{
			*ToUnstructured(Deployment("test", WithDeploymentContainer(Container("test-name", "gcr.io/test/test-image:explicit-tag")))),
		},
		want: "explicit-tag",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest, err := mf.ManifestFrom(test.input)

			if err != nil {
				t.Errorf("mf.ManifestFrom error: %v", err)
			}

			got := VersionFromDeploymentImage(&manifest, "test-default")
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("VersionFromDeploymentImage (-want, +got) = %v", diff)
			}
		})
	}
}
