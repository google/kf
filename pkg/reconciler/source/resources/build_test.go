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
)

func ExampleBuildName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Generation = 5

	fmt.Println(BuildName(source))

	// Output: my-source-5
}

func ExampleAppImageName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Generation = 2

	fmt.Println(AppImageName(source))
	// Output: app-my-namespace-my-source:2
}

func ExampleJoinRepositoryImage() {
	fmt.Println(JoinRepositoryImage("repo", "app"))

	// Output: repo/app
}

func ExampleMakeBuild() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Generation = 5
	source.Spec.BuildpackBuild.Source = "some-source"
	source.Spec.BuildpackBuild.Registry = "some-registry"

	build, err := MakeBuild(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", BuildName(source))
	fmt.Println("Label Count:", len(build.Labels))
	fmt.Println("Managed By:", build.Labels[managedByLabel])
	fmt.Println("Arg Count:", len(build.Spec.Template.Arguments))
	fmt.Println("Output Image:", build.Spec.Template.Arguments[0].Value)

	// Output: Name: my-source-5
	// Label Count: 1
	// Managed By: kf
	// Arg Count: 2
	// Output Image: some-registry/app-my-namespace-my-source:5
}
