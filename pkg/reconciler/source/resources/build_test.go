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

func ExampleBuildSecretName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"

	fmt.Println(BuildSecretName(source))

	// Output: my-source
}

func exampleSource() *v1alpha1.Source {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Spec.ServiceAccount = "some-account"
	source.Spec.BuildpackBuild.Source = "some-source"
	source.Spec.BuildpackBuild.Image = "gcr.io/image:123"
	source.Spec.BuildpackBuild.Stack = "gcr.io/kf-releases/run:latest"
	source.Spec.BuildpackBuild.BuildpackBuilder = "some-buildpack-builder"
	source.Spec.BuildpackBuild.Env = []corev1.EnvVar{
		{
			Name:  "FOO",
			Value: "bar",
		},
		{
			Name:  "BAR",
			Value: "baz",
		},
	}
	return source
}

func ExampleMakeBuild_build() {
	source := exampleSource()

	build, _, err := MakeBuild(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", build.Name)
	fmt.Println("Label Count:", len(build.Labels))
	fmt.Println("Managed By:", build.Labels[managedByLabel])
	fmt.Println("OwnerReferences Count:", len(build.OwnerReferences))
	fmt.Println("Arg Count:", len(build.Spec.Template.Arguments))
	fmt.Println("Output Image:", v1alpha1.GetBuildArg(build, v1alpha1.BuildArgImage))
	fmt.Println("Stack:", v1alpha1.GetBuildArg(build, v1alpha1.BuildArgBuildpackRunImage))
	fmt.Println("Env Count:", len(build.Spec.Template.Env))
	fmt.Println("Env.Name:", build.Spec.Template.Env[0].Name)
	fmt.Println("Env.ValueFrom.SecretKeyRef.LocalObjectReference.Name:", build.Spec.Template.Env[0].ValueFrom.SecretKeyRef.LocalObjectReference.Name)
	fmt.Println("Env.ValueFrom.SecretKeyRef.Key:", build.Spec.Template.Env[0].ValueFrom.SecretKeyRef.Key)

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// OwnerReferences Count: 1
	// Arg Count: 4
	// Output Image: gcr.io/image:123
	// Stack: gcr.io/kf-releases/run:latest
	// Env Count: 2
	// Env.Name: FOO
	// Env.ValueFrom.SecretKeyRef.LocalObjectReference.Name: my-source
	// Env.ValueFrom.SecretKeyRef.Key: FOO
}

func ExampleMakeBuild_secret() {
	source := exampleSource()

	_, secret, err := MakeBuild(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", secret.Name)
	fmt.Println("Label Count:", len(secret.Labels))
	fmt.Println("Managed By:", secret.Labels[managedByLabel])
	fmt.Println("OwnerReferences Count:", len(secret.OwnerReferences))
	fmt.Println("Data['BAR']:", string(secret.Data["BAR"]))
	fmt.Println("Data['FOO']:", string(secret.Data["FOO"]))

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// OwnerReferences Count: 1
	// Data['BAR']: baz
	// Data['FOO']: bar
}
