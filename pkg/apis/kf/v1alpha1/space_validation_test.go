// Copyright 2019 Google LLC
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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestSpaceValidation(t *testing.T) {
	goodBuildpackBuild := SpaceSpecBuildpackBuild{
		BuilderImage:      DefaultBuilderImage,
		ContainerRegistry: "gcr.io/test",
	}
	goodExecuton := SpaceSpecExecution{
		Domains: []SpaceDomain{{Domain: "example.com", Default: true}},
	}

	goodSpaceSpec := SpaceSpec{
		BuildpackBuild: goodBuildpackBuild,
		Execution:      goodExecuton,
	}

	cases := map[string]struct {
		space *Space
		want  *apis.FieldError
	}{
		"good": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec:       goodSpaceSpec,
			},
		},
		"missing name": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: ""},
				Spec:       goodSpaceSpec,
			},
			want: apis.ErrMissingField("name"),
		},
		"reserved name: kf": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "kf"},
				Spec:       goodSpaceSpec,
			},
			want: apis.ErrInvalidValue("kf", "name"),
		},
		"reserved name: default": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
				Spec:       goodSpaceSpec,
			},
			want: apis.ErrInvalidValue("default", "name"),
		},
		"no container registry": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					Execution: goodExecuton,
					BuildpackBuild: SpaceSpecBuildpackBuild{
						BuilderImage: DefaultBuilderImage,
					},
				},
			},
			want: apis.ErrMissingField("spec.buildpackBuild.containerRegistry"),
		},
		"no builder image": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					Execution: goodExecuton,
					BuildpackBuild: SpaceSpecBuildpackBuild{
						ContainerRegistry: "gcr.io/test",
					},
				},
			},
			want: apis.ErrMissingField("spec.buildpackBuild.builderImage"),
		},
		"no domains": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					BuildpackBuild: SpaceSpecBuildpackBuild{
						ContainerRegistry: "gcr.io/test",
						BuilderImage:      DefaultBuilderImage,
					},
				},
			},
			want: apis.ErrMissingField("spec.execution.domains"),
		},
		"no default domain": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					Execution: SpaceSpecExecution{
						Domains: []SpaceDomain{
							{Domain: "example.com"},
						},
					},
					BuildpackBuild: SpaceSpecBuildpackBuild{
						ContainerRegistry: "gcr.io/test",
						BuilderImage:      DefaultBuilderImage,
					},
				},
			},
			want: apis.ErrInvalidArrayValue(SpaceDomain{Domain: "example.com"}, "spec.execution.domains", 0),
		},
		"multiple default domains": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					Execution: SpaceSpecExecution{
						Domains: []SpaceDomain{
							{Domain: "example.com", Default: true},
							{Domain: "other-example.com", Default: true},
						},
					},
					BuildpackBuild: SpaceSpecBuildpackBuild{
						ContainerRegistry: "gcr.io/test",
						BuilderImage:      DefaultBuilderImage,
					},
				},
			},
			want: apis.ErrInvalidArrayValue(SpaceDomain{Domain: "other-example.com", Default: true}, "spec.execution.domains", 1),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.space.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
