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

package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"
	logtesting "knative.dev/pkg/logging/testing"
)

func TestBuildSpecBuilders(t *testing.T) {

	cases := map[string]struct {
		buildSpec BuildSpec
	}{
		"BuildpackV2Build single buildpack": {
			buildSpec: BuildpackV2Build(
				"some/source/image",
				config.StackV2Definition{
					Name:  "stack-name",
					Image: "google/base:latest",
				},
				[]string{"buildpack-1"},
				true,
			),
		},
		"BuildpackV2Build multiple buildpack": {
			buildSpec: BuildpackV2Build(
				"some/source/image",
				config.StackV2Definition{
					Name:  "stack-name",
					Image: "google/base:latest",
					NodeSelector: map[string]string{
						"disktype": "ssd",
						"cpu":      "amd64",
					},
				},
				[]string{"buildpack-1", "buildpack-2"},
				false,
			),
		},
		"BuildpackV3Build": {
			buildSpec: BuildpackV3Build(
				"some/source/image",
				config.StackV3Definition{
					Name:       "stack-name",
					BuildImage: "build/image:latest",
					RunImage:   "run/image:latest",
					NodeSelector: map[string]string{
						"disktype": "ssd",
						"cpu":      "amd64",
					},
				},
				[]string{"buildpack-1", "buildpack-2"},
			),
		},
		"DockerfileBuild": {
			buildSpec: DockerfileBuild("some/source/image", "path/to/Dockerfile"),
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// Test validity
			errs := tc.buildSpec.Validate(store.ToContext(context.Background()))
			testutil.AssertEqual(t, "empty error", (*apis.FieldError)(nil), errs)

			// Check serialization
			testutil.AssertGoldenJSON(t, "buildSpec", tc.buildSpec)
		})
	}
}

func TestBuildParam_ToTektonParam(t *testing.T) {
	// the over the wire format should be compatible with Tekton
	raw := `{"name": "some-name", "value": "some-value"}`

	var wantTekton tektonv1beta1.Param
	if err := json.Unmarshal([]byte(raw), &wantTekton); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bp := BuildParam{}
	if err := json.Unmarshal([]byte(raw), &bp); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	testutil.AssertEqual(t, "Tekton Param", bp.ToTektonParam(), wantTekton)

	// Sanity check
	testutil.AssertEqual(t, "Name", bp.Name, wantTekton.Name)
	testutil.AssertEqual(t, "Value", bp.Value, wantTekton.Value.StringVal)
}
