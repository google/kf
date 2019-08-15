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

package resources

import (
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func ExampleBuildName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"

	fmt.Println(BuildName(source))

	// Output: my-source
}

func ExampleMakeBuild() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Spec.ServiceAccount = "some-account"
	source.Spec.BuildpackBuild.Source = "some-source"
	source.Spec.BuildpackBuild.Image = "gcr.io/image:123"
	source.Spec.BuildpackBuild.BuildpackBuilder = "some-buildpack-builder"
	source.Spec.BuildpackBuild.Env = []corev1.EnvVar{
		{
			Name:  "some",
			Value: "variable",
		},
	}

	build, err := MakeBuild(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", BuildName(source))
	fmt.Println("Label Count:", len(build.Labels))
	fmt.Println("Managed By:", build.Labels[managedByLabel])
	fmt.Println("Service Account:", build.Spec.ServiceAccountName)
	fmt.Println("Arg Count:", len(build.Spec.Template.Arguments))
	fmt.Println("Output Image:", build.Spec.Template.Arguments[0].Value)
	fmt.Println("Env:", build.Spec.Template.Env[0].Name, "=", build.Spec.Template.Env[0].Value)

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// Service Account: some-account
	// Arg Count: 3
	// Output Image: gcr.io/image:123
	// Env: some = variable
}
