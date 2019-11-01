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

	fmt.Println(TaskRunName(source))

	// Output: my-source
}

func ExampleBuildSecretName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"

	fmt.Println(TaskRunSecretName(source))

	// Output: my-source
}

func exampleBuildpackSource() *v1alpha1.Source {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Spec.ServiceAccount = "some-account"
	source.Spec.BuildpackBuild.Source = "gcr.io/image:456"
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

func exampleContainerSource() *v1alpha1.Source {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Spec.ServiceAccount = "some-account"
	source.Spec.ContainerImage.Image = "gcr.io/image:123"
	return source
}

func exampleDockerSource() *v1alpha1.Source {
	source := &v1alpha1.Source{}
	source.Name = "my-source"
	source.Namespace = "my-namespace"
	source.Spec.ServiceAccount = "some-account"
	source.Spec.Dockerfile.Source = "gcr.io/image:456"
	source.Spec.Dockerfile.Image = "gcr.io/image:123"
	return source
}

func ExampleMakeTaskRun_buildpack_taskRun() {
	source := exampleBuildpackSource()

	taskRun, _, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", taskRun.Name)
	fmt.Println("Label Count:", len(taskRun.Labels))
	fmt.Println("Managed By:", taskRun.Labels[managedByLabel])
	fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName)
	fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences))
	fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params))
	fmt.Println("Input Image:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamSourceContainer))
	fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources))
	fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL))
	fmt.Println("Stack:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamBuildpackRunImage))
	fmt.Println("Env Name:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamEnvSecret))

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// Service Account: some-account
	// OwnerReferences Count: 1
	// Input Count: 5
	// Input Image: gcr.io/image:456
	// Output Count: 1
	// Output Image: gcr.io/image:123
	// Stack: gcr.io/kf-releases/run:latest
	// Env Name: my-source
}

func ExampleMakeTaskRun_buildpack_secret() {
	source := exampleBuildpackSource()

	_, secret, err := MakeTaskRun(source)
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

func ExampleMakeTaskRun_container_taskRun() {
	source := exampleContainerSource()

	taskRun, _, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", taskRun.Name)
	fmt.Println("Label Count:", len(taskRun.Labels))
	fmt.Println("Managed By:", taskRun.Labels[managedByLabel])
	fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName)
	fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences))
	fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params))
	fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources))
	fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL))

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// Service Account: some-account
	// OwnerReferences Count: 1
	// Input Count: 0
	// Output Count: 1
	// Output Image: gcr.io/image:123
}

func ExampleMakeTaskRun_container_secret() {
	source := exampleContainerSource()

	_, secret, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Is nil:", secret == nil)

	// Output: Is nil: true
}

func ExampleMakeTaskRun_docker_taskRun() {
	source := exampleDockerSource()

	taskRun, _, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", taskRun.Name)
	fmt.Println("Label Count:", len(taskRun.Labels))
	fmt.Println("Managed By:", taskRun.Labels[managedByLabel])
	fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName)
	fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences))
	fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params))
	fmt.Println("Input Image:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamSourceContainer))
	fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources))
	fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL))

	// Output: Name: my-source
	// Label Count: 1
	// Managed By: kf
	// Service Account: some-account
	// OwnerReferences Count: 1
	// Input Count: 2
	// Input Image: gcr.io/image:456
	// Output Count: 1
	// Output Image: gcr.io/image:123
}

func ExampleMakeTaskRun_docker_secret() {
	source := exampleDockerSource()

	_, secret, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	fmt.Println("Is nil:", secret == nil)

	// Output: Is nil: true
}
