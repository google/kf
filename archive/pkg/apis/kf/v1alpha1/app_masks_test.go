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

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAppSpecSourceMask(t *testing.T) {
	want := SourceSpec{
		UpdateRequests: 10,
		ServiceAccount: "",
		BuildpackBuild: SourceSpecBuildpackBuild{
			Buildpack:        "custom-buildpack",
			BuildpackBuilder: "",
			Env:              []corev1.EnvVar{{Name: "env-key", Value: "env-value"}},
			Image:            "",
			Source:           "gcr.io/custom-source:mysource",
			Stack:            "cflinuxfs3",
		},
		ContainerImage: SourceSpecContainerImage{
			Image: "mysql/mysql",
		},
		Dockerfile: SourceSpecDockerfile{
			Image:  "",
			Path:   "path/to/Dockerfile",
			Source: "gcr.io/custom-source:dockerfilesource",
		},
	}

	input := SourceSpec{
		UpdateRequests: 10,
		ServiceAccount: "custom-sa",
		BuildpackBuild: SourceSpecBuildpackBuild{
			Buildpack:        "custom-buildpack",
			BuildpackBuilder: "custom-builder",
			Env:              []corev1.EnvVar{{Name: "env-key", Value: "env-value"}},
			Image:            "gcr.io/custom-image:label",
			Source:           "gcr.io/custom-source:mysource",
			Stack:            "cflinuxfs3",
		},
		ContainerImage: SourceSpecContainerImage{
			Image: "mysql/mysql",
		},
		Dockerfile: SourceSpecDockerfile{
			Image:  "gcr.io/custom-image:label",
			Path:   "path/to/Dockerfile",
			Source: "gcr.io/custom-source:dockerfilesource",
		},
	}

	actual := AppSpecSourceMask(input)

	testutil.AssertEqual(t, "masked values", want, actual)
}
