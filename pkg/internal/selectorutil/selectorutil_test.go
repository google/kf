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

package selectorutil_test

import (
	"testing"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/selectorutil"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func Test_GetNodeSelector(t *testing.T) {
	tests := map[string]struct {
		app   *v1alpha1.App
		space *v1alpha1.Space
		want  map[string]string
	}{
		"default": {
			app:   &v1alpha1.App{},
			space: &v1alpha1.Space{},
			want:  nil,
		},
		"populated_stack_overwrite": {
			app: &v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Build: v1alpha1.AppSpecBuild{
						Spec: &v1alpha1.BuildSpec{
							NodeSelector: map[string]string{
								"disktype": "ssd10",
							},
						},
					},
				},
			},
			space: &v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: map[string]string{
							"disktype": "ssd",
							"cpu":      "amd64",
						},
					},
				},
			},

			want: map[string]string{
				"disktype": "ssd10",
				"cpu":      "amd64",
			},
		},
		"populated_nooverwrite": {
			app: &v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Build: v1alpha1.AppSpecBuild{},
				},
			},
			space: &v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: map[string]string{
							"disktype": "ssd",
							"cpu":      "amd64",
						},
					},
				},
			},

			want: map[string]string{
				"disktype": "ssd",
				"cpu":      "amd64",
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {

			got := selectorutil.GetNodeSelector(tc.app.Spec.Build.Spec, tc.space)
			testutil.AssertEqual(t, "GetNodeSelector", tc.want, got)
		})
	}
}
