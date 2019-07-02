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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestApp_Validate(t *testing.T) {
	goodInstances := AppSpecInstances{Stopped: true}
	goodTemplate := AppSpecTemplate{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{}},
		},
	}

	cases := map[string]struct {
		spec App
		want *apis.FieldError
	}{
		"valid": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
				},
			},
		},
		"invalid instances": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: AppSpecInstances{Exactly: intPtr(-1)},
				},
			},
			want: apis.ErrInvalidValue(-1, "spec.instances.exactly"),
		},
		"invalid template": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  AppSpecTemplate{},
					Instances: goodInstances,
				},
			},
			want: apis.ErrMissingField("spec.template.spec.containers"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}

}

func TestAppSpecInstances_Validate(t *testing.T) {
	// These test cases are broken out separately because they're
	// too extenstive to copy the whole service struct for.

	cases := map[string]struct {
		spec AppSpecInstances
		want *apis.FieldError
	}{
		"blank": {
			spec: AppSpecInstances{},
		},
		"stopped": {
			spec: AppSpecInstances{Stopped: true},
		},
		"valid minmax": {
			spec: AppSpecInstances{Min: intPtr(3), Max: intPtr(5)},
		},
		"valid exactly": {
			spec: AppSpecInstances{Exactly: intPtr(3)},
		},
		"exactly and min": {
			spec: AppSpecInstances{Exactly: intPtr(3), Min: intPtr(3)},
			want: apis.ErrMultipleOneOf("exactly", "min"),
		},
		"exactly and max": {
			spec: AppSpecInstances{Exactly: intPtr(3), Max: intPtr(3)},
			want: apis.ErrMultipleOneOf("exactly", "max"),
		},
		"exactly lt 0": {
			spec: AppSpecInstances{Exactly: intPtr(-1)},
			want: apis.ErrInvalidValue(-1, "exactly"),
		},
		"min lt 0": {
			spec: AppSpecInstances{Min: intPtr(-1)},
			want: apis.ErrInvalidValue(-1, "min"),
		},
		"max lt 0": {
			spec: AppSpecInstances{Max: intPtr(-1)},
			want: apis.ErrInvalidValue(-1, "max"),
		},
		"max lt min": {
			spec: AppSpecInstances{Max: intPtr(1), Min: intPtr(50)},
			want: &apis.FieldError{Message: "max must be >= min", Paths: []string{"min", "max"}},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestValidatePodSpec(t *testing.T) {
	cases := map[string]struct {
		spec corev1.PodSpec
		want *apis.FieldError
	}{
		"missing container": {
			spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
			want: apis.ErrMissingField("containers"),
		},
		"too many containers": {
			spec: corev1.PodSpec{
				Containers: []corev1.Container{{}, {}},
			},
			want: apis.ErrMultipleOneOf("containers"),
		},
		"container has image": {
			spec: corev1.PodSpec{
				Containers: []corev1.Container{{Image: "some-image"}},
			},
			want: apis.ErrDisallowedFields("image"),
		},
		"upstream failure": {
			// NOTE: this test is intended to show that a Knative Serving error will
			// be passed thorugh, it doesn't matter which upstream error. In the
			// future Knative Serving may decide to allow InitContainers in which case
			// this test will need to choose some other invalid field.
			spec: corev1.PodSpec{
				Containers:     []corev1.Container{{}},
				InitContainers: []corev1.Container{{Image: "some-image"}},
			},
			want: apis.ErrDisallowedFields("initContainers"),
		},
		"missing image is okay": {
			spec: corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := ValidatePodSpec(tc.spec)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
