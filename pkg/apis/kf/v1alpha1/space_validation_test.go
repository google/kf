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
	goodSpaceSpec := SpaceSpec{}

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
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.space.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})

	}
}
