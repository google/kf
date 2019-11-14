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

func ExampleTaskRunName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"

	if _, err := fmt.Println(TaskRunName(source)); err != nil {
		panic(err)
	}

	// Output: my-source
}

func ExampleTaskRunSecretName() {
	source := &v1alpha1.Source{}
	source.Name = "my-source"

	if _, err := fmt.Println(TaskRunSecretName(source)); err != nil {
		panic(err)
	}

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

	if _, err := fmt.Println("Name:", taskRun.Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Label Count:", len(taskRun.Labels)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Managed By:", taskRun.Labels[managedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Input Image:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamSourceContainer)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Stack:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamBuildpackRunImage)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Env Name:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamEnvSecret)); err != nil {
		panic(err)
	}

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

	if _, err := fmt.Println("Name:", secret.Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Label Count:", len(secret.Labels)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Managed By:", secret.Labels[managedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("OwnerReferences Count:", len(secret.OwnerReferences)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Data['BAR']:", string(secret.Data["BAR"])); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Data['FOO']:", string(secret.Data["FOO"])); err != nil {
		panic(err)
	}

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

	if _, err := fmt.Println("Name:", taskRun.Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Label Count:", len(taskRun.Labels)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Managed By:", taskRun.Labels[managedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL)); err != nil {
		panic(err)
	}

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

	if _, err := fmt.Println("Is nil:", secret == nil); err != nil {
		panic(err)
	}

	// Output: Is nil: true
}

func ExampleMakeTaskRun_docker_taskRun() {
	source := exampleDockerSource()

	taskRun, _, err := MakeTaskRun(source)
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Println("Name:", taskRun.Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Label Count:", len(taskRun.Labels)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Managed By:", taskRun.Labels[managedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Input Count:", len(taskRun.Spec.Inputs.Params)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Input Image:", v1alpha1.GetTaskRunInputParam(taskRun, v1alpha1.TaskRunParamSourceContainer)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Count:", len(taskRun.Spec.Outputs.Resources)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Output Image:", v1alpha1.GetTaskRunOutputResource(taskRun, v1alpha1.TaskRunResourceNameImage, v1alpha1.TaskRunResourceURL)); err != nil {
		panic(err)
	}

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

	if _, err := fmt.Println("Is nil:", secret == nil); err != nil {
		panic(err)
	}

	// Output: Is nil: true
}
