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
	goodBuildpackBuild := SourceSpecBuildpackBuild{
		Source:           "some-source-image",
		Buildpack:        "some-buildpack",
		Stack:            "some-stack",
		BuildpackBuilder: "some-buildpack-builder",
		Image:            "some-container-registry",
	}
	badBuildpackBuild := SourceSpecBuildpackBuild{
		Source:           "missing-stack",
		BuildpackBuilder: "no-stack",
		Image:            "still-no-stack",
	}
	goodContainerImage := SourceSpecContainerImage{
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
			want: apis.ErrMultipleOneOf("spec.buildpackBuild", "spec.containerImage", "spec.dockerfile"),
		},
		"invalid neither": {
			spec: Source{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourceSpec{},
			},
			want: apis.ErrMissingOneOf("spec.buildpackBuild", "spec.containerImage", "spec.dockerfile"),
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

func TestSourceSpecBuildpackBuild_Validate(t *testing.T) {
	cases := map[string]struct {
		spec SourceSpecBuildpackBuild
		want *apis.FieldError
	}{
		"valid": {
			spec: SourceSpecBuildpackBuild{
				Source:           "some-image",
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Image:            "some-registry",
			},
		},
		"missing source": {
			spec: SourceSpecBuildpackBuild{
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Image:            "some-registry",
			},
			want: apis.ErrMissingField("source"),
		},
		"missing stack": {
			spec: SourceSpecBuildpackBuild{
				Source:           "some-image",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
				Image:            "some-registry",
			},
			want: apis.ErrMissingField("stack"),
		},
		"missing buildpackBuilder": {
			spec: SourceSpecBuildpackBuild{
				Source:    "some-image",
				Stack:     "some-stack",
				Buildpack: "some-buildpack",
				Image:     "some-registry",
			},
			want: apis.ErrMissingField("buildpackBuilder"),
		},
		"missing image": {
			spec: SourceSpecBuildpackBuild{
				Source:           "some-image",
				Stack:            "some-stack",
				Buildpack:        "some-buildpack",
				BuildpackBuilder: "buildpackBuilder",
			},
			want: apis.ErrMissingField("image"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestSourceSpecDockerfile_Valdiate(t *testing.T) {
	t.Errorf("TODO")
}
