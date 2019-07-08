// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestSource_Validate(t *testing.T) {
	goodBuildpackBuild := AppSpecSourceBuildpackBuild{
		Source:           "some-source-image",
		Buildpack:        "some-buildpack",
		Stack:            "some-stack",
		BuildpackBuilder: "some-buildpack-builder",
		Registry:         "some-container-registry",
	}
	badBuildpackBuild := AppSpecSourceBuildpackBuild{
		Source:           "missing-stack",
		BuildpackBuilder: "no-stack",
		Registry:         "still-no-stack",
	}
	goodContainerImage := AppSpecSourceContainerImage{
		Image: "some-container-image",
	}

	cases := map[string]struct {
		spec Source
		want *apis.FieldError
	}{
		"valid buildpackBuild": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{
					BuildpackBuild: goodBuildpackBuild,
				},
			},
		},
		"valid containerImage": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{
					ContainerImage: goodContainerImage,
				},
			},
		},
		"invalid both": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{
					ContainerImage: goodContainerImage,
					BuildpackBuild: goodBuildpackBuild,
				},
			},
			want: apis.ErrMultipleOneOf("spec.buildpackBuild", "spec.containerImage"),
		},
		"invalid neither": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{},
			},
			want: apis.ErrMissingOneOf("spec.buildpackBuild", "spec.containerImage"),
		},
		"invalid buildpackBuild": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{
					BuildpackBuild: badBuildpackBuild,
				},
			},
			want: apis.ErrMissingField("spec.stack"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}

}

func TestAppSpecSourceBuildpackBuild_Validate(t *testing.T) {
	cases := map[string]struct {
		spec AppSpecSourceBuildpackBuild
		want *apis.FieldError
	}{
		"valid": {
			spec: AppSpecSourceBuildpackBuild{
				Source:           "some-image",
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Registry:         "some-registry",
			},
		},
		"missing image": {
			spec: AppSpecSourceBuildpackBuild{
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Registry:         "some-registry",
			},
			want: apis.ErrMissingField("source"),
		},
		"missing stack": {
			spec: AppSpecSourceBuildpackBuild{
				Source:           "some-image",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Registry:         "some-registry",
			},
			want: apis.ErrMissingField("stack"),
		},
		"missing buildpackBuilder": {
			spec: AppSpecSourceBuildpackBuild{
				Source:    "some-image",
				Stack:     "some-stack",
				Buildpack: "some-buildpack",
				Registry:  "some-registry",
			},
			want: apis.ErrMissingField("buildpackBuilder"),
		},
		"missing registry": {
			spec: AppSpecSourceBuildpackBuild{
				Source:           "some-image",
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
			},
			want: apis.ErrMissingField("registry"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
