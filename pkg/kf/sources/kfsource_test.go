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

	fmt.Println("Name:", source.GetName())
	fmt.Println("Namespace:", source.GetNamespace())
	fmt.Println("Source:", source.GetBuildpackBuildSource())
	fmt.Println("Buildpack:", source.GetBuildpackBuildBuildpack())
	fmt.Println("Image:", source.GetBuildpackBuildImage())

	for _, env := range source.GetBuildpackBuildEnv() {
		fmt.Println("Env:", env.Name, "=", env.Value)
	}

	// Output: Name: my-buildpack-build
	// Namespace: my-namespace
	// Source: gcr.io/my-source-code-image
	// Buildpack: java
	// Image: gcr.io/some-registry/my-image:latest
	// Env: JAVA_VERSION = 11
}

func ExampleKfSource_docker() {
	source := NewKfSource()

	source.SetName("my-docker-build")
	source.SetNamespace("my-namespace")
	source.SetContainerImageSource("mysql/mysql")

	fmt.Println("Name:", source.GetName())
	fmt.Println("Namespace:", source.GetNamespace())
	fmt.Println("Source:", source.GetContainerImageSource())

	// Output: Name: my-docker-build
	// Namespace: my-namespace
	// Source: mysql/mysql
}
