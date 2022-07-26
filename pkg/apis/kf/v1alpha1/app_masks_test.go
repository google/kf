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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAppSpecBuildMask(t *testing.T) {
	want := BuildSpec{
		SourcePackage: corev1.LocalObjectReference{
			Name: "some-name",
		},
		BuildTaskRef: buildpackV3BuildTaskRef(),
		Params: []BuildParam{
			{
				Name: "IMAGE",
			},
			{
				Name: "NOT_IMAGE",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "key",
				Value: "val",
			},
		},
	}

	input := BuildSpec{
		SourcePackage: corev1.LocalObjectReference{
			Name: "some-name",
		},
		BuildTaskRef: buildpackV3BuildTaskRef(),
		Params: []BuildParam{
			{
				Name: "IMAGE",
			},
			{
				Name: "NOT_IMAGE",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "key",
				Value: "val",
			},
		},
	}

	actual := AppSpecBuildMask(input)

	testutil.AssertEqual(t, "masked values", want, actual)
}
