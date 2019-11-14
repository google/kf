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

package sources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func ExampleKfSource_buildpack() {
	source := NewKfSource()

	source.SetName("my-buildpack-build")
	source.SetNamespace("my-namespace")
	source.SetBuildpackBuildSource("gcr.io/my-source-code-image")
	source.SetBuildpackBuildEnv([]corev1.EnvVar{{Name: "JAVA_VERSION", Value: "11"}})
	source.SetBuildpackBuildBuildpack("java")
	source.SetBuildpackBuildImage("gcr.io/some-registry/my-image:latest")
	source.SetBuildpackBuildStack("cflinuxfs3")

	if _, err := fmt.Println("Name:", source.GetName()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Namespace:", source.GetNamespace()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Source:", source.GetBuildpackBuildSource()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Buildpack:", source.GetBuildpackBuildBuildpack()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Image:", source.GetBuildpackBuildImage()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Stack:", source.GetBuildpackBuildStack()); err != nil {
		panic(err)
	}

	for _, env := range source.GetBuildpackBuildEnv() {
		if _, err := fmt.Println("Env:", env.Name, "=", env.Value); err != nil {
			panic(err)
		}
	}

	// Output: Name: my-buildpack-build
	// Namespace: my-namespace
	// Source: gcr.io/my-source-code-image
	// Buildpack: java
	// Image: gcr.io/some-registry/my-image:latest
	// Stack: cflinuxfs3
	// Env: JAVA_VERSION = 11
}

func ExampleKfSource_docker() {
	source := NewKfSource()

	source.SetName("my-docker-build")
	source.SetNamespace("my-namespace")
	source.SetContainerImageSource("mysql/mysql")

	if _, err := fmt.Println("Name:", source.GetName()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Namespace:", source.GetNamespace()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Source:", source.GetContainerImageSource()); err != nil {
		panic(err)
	}

	// Output: Name: my-docker-build
	// Namespace: my-namespace
	// Source: mysql/mysql
}
