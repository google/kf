// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func ExampleTaskRunName() {
	task := &v1alpha1.Task{}
	task.Name = "my-task"

	fmt.Println(TaskRunName(task))

	// Output: my-task
}

func exampleCustomTask() (*v1alpha1.Task, *v1alpha1.App) {
	task := &v1alpha1.Task{}
	task.Name = "my-task"
	task.UID = "0d5c53ff-edf1-4d42-8d1a-fdd5b5cf23d3"
	task.Namespace = "my-namespace"
	task.Spec.AppRef = corev1.LocalObjectReference{
		Name: "my-app",
	}

	app := &v1alpha1.App{
		Status: v1alpha1.AppStatus{
			ServiceAccountName: "my-sa",
		},
	}

	return task, app
}

func exampleSpace() *v1alpha1.Space {
	return &v1alpha1.Space{
		Status: v1alpha1.SpaceStatus{
			BuildConfig: v1alpha1.SpaceStatusBuildConfig{
				ServiceAccount: "space-account",
			},
		},
	}
}

func ExampleMakeTaskRun_verifyTaskRunSetup() {
	cfg := config.BuiltinDefaultsConfig()

	task, app := exampleCustomTask()

	space := exampleSpace()

	taskRun, err := MakeTaskRun(cfg, task, app, space, []string{"some-command"})

	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", taskRun.Name)
	fmt.Println("Label Count:", len(taskRun.Labels))
	fmt.Println("Managed By:", taskRun.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("NetworkPolicy:", taskRun.Labels[v1alpha1.NetworkPolicyLabel])
	fmt.Println("Service account:", taskRun.Spec.ServiceAccountName)
	fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences))

	// Output: Name: my-task
	// Label Count: 5
	// Managed By: kf
	// NetworkPolicy: app
	// Service account: my-sa
	// OwnerReferences Count: 1
}

func ExampleMakeTaskRun_verifyAppArgsByDefault() {
	cfg := config.BuiltinDefaultsConfig()

	task, app := exampleCustomTask()

	app.Spec = v1alpha1.AppSpec{
		Template: v1alpha1.AppSpecTemplate{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Args: []string{"app-command"},
					},
				},
			},
		},
	}

	space := exampleSpace()

	taskRun, err := MakeTaskRun(cfg, task, app, space, []string{"some-command"})

	if err != nil {
		panic(err)
	}

	args := taskRun.Spec.TaskSpec.Steps[0].Args
	fmt.Println("user-container Args:", args)

	// Output: user-container Args: [app-command]
}

func ExampleMakeTaskRun_verifyTaskParamsOverrides() {
	cfg := config.BuiltinDefaultsConfig()

	task, app := exampleCustomTask()

	space := exampleSpace()

	task.Spec.CPU = "2"
	task.Spec.Memory = "2G"
	task.Spec.Disk = "2G"
	task.Spec.Command = "task-command"

	taskRun, err := MakeTaskRun(cfg, task, app, space, []string{"some-command"})

	if err != nil {
		panic(err)
	}

	step := taskRun.Spec.TaskSpec.Steps[0]

	cpu := step.Resources.Requests[corev1.ResourceCPU]
	memory := step.Resources.Requests[corev1.ResourceMemory]
	disk := step.Resources.Requests[corev1.ResourceEphemeralStorage]

	fmt.Println("user-container CPU:", cpu.String())
	fmt.Println("user-container Memory:", memory.String())
	fmt.Println("user-container Disk:", disk.String())
	fmt.Println("user-container Args:", step.Args)
	fmt.Println("user-container Entrypoint:", step.Command)

	// Output: user-container CPU: 2
	// user-container Memory: 2G
	// user-container Disk: 2G
	// user-container Args: [task-command]
	// user-container Entrypoint: [some-command]
}

func ExampleMakeTaskRun_verifyAppEntrypointIsUsed() {
	cfg := config.BuiltinDefaultsConfig()

	task, app := exampleCustomTask()

	app.Spec = v1alpha1.AppSpec{
		Template: v1alpha1.AppSpecTemplate{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Command: []string{"app-entrypoint"},
					},
				},
			},
		},
	}

	space := exampleSpace()

	taskRun, err := MakeTaskRun(cfg, task, app, space, []string{"some-command"})
	if err != nil {
		panic(err)
	}

	command := taskRun.Spec.TaskSpec.Steps[0].Command
	fmt.Println("Container entrypoint:", command)

	// Output: Container entrypoint: [app-entrypoint]
}

func ExampleMakeTaskRun_verifyTaskRunCancelled() {
	cfg := config.BuiltinDefaultsConfig()

	task, app := exampleCustomTask()

	task.Spec.Terminated = true

	space := exampleSpace()

	taskRun, err := MakeTaskRun(cfg, task, app, space, []string{"some-command"})

	if err != nil {
		panic(err)
	}

	fmt.Println("TaskRun status:", taskRun.Spec.Status)

	// Output: TaskRun status: TaskRunCancelled
}
