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
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestTask_Validate(t *testing.T) {
	goodSpec := TaskSpec{
		AppRef: corev1.LocalObjectReference{
			Name: "appRef",
		},
	}

	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	cases := map[string]struct {
		spec Task
		want *apis.FieldError
	}{
		"valid spec": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: goodSpec,
			},
		},
		"invalid ObjectMeta": {
			spec: Task{
				ObjectMeta: badMeta,
				Spec:       goodSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"missing appRef": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
			},
			want: apis.ErrMissingField("spec.appRef"),
		},
		"invalid cpu": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: TaskSpec{
					AppRef: corev1.LocalObjectReference{
						Name: "appRef",
					},
					CPU: "invalid",
				},
			},
			want: apis.ErrInvalidValue("invalid", "spec.cpu"),
		},
		"invalid memory": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: TaskSpec{
					AppRef: corev1.LocalObjectReference{
						Name: "appRef",
					},
					Memory: "invalid",
				},
			},
			want: apis.ErrInvalidValue("invalid", "spec.memory"),
		},
		"invalid disk": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: TaskSpec{
					AppRef: corev1.LocalObjectReference{
						Name: "appRef",
					},
					Disk: "invalid",
				},
			},
			want: apis.ErrInvalidValue("invalid", "spec.disk"),
		},
		"multi params invalid": {
			spec: Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: TaskSpec{
					AppRef: corev1.LocalObjectReference{
						Name: "appRef",
					},
					CPU:    "invalid",
					Memory: "invalid",
					Disk:   "invalid",
				},
			},
			want: apis.ErrInvalidValue("invalid", "spec.cpu, spec.disk, spec.memory"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
