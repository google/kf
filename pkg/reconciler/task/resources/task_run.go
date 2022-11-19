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
	"strings"
	"time"

	"github.com/alessio/shellescape"
	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appresources "github.com/google/kf/v2/pkg/reconciler/app/resources"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

const (
	tektonPipelineTaskLabel = "tekton.dev/pipelineTask"
)

// TaskRunName gets the name of a TaskRun for a Kf Task.
func TaskRunName(task *v1alpha1.Task) string {
	return task.Name
}

func makeObjectMeta(name string, task *v1alpha1.Task, app *v1alpha1.App) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: task.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(task),
		},
		// Copy labels from the parent.
		Labels: v1alpha1.UnionMaps(
			app.GetLabels(),
			app.ComponentLabels(v1alpha1.TaskComponentName),
			task.GetLabels(),
			map[string]string{
				v1alpha1.ManagedByLabel:     "kf",
				v1alpha1.NetworkPolicyLabel: v1alpha1.NetworkPolicyApp,

				// NOTE: This label is used by the CLI to find relevant pods
				// for logging.
				tektonPipelineTaskLabel: v1alpha1.DefaultUserContainerName,
			}),
		Annotations: map[string]string{
			// Allow Istio injection on Tekton tasks.
			"sidecar.istio.io/inject": "true",
		},
	}
}

// MakeTaskRun creates a TaskRun for the given Kf Task in the given Space.
func MakeTaskRun(
	cfg *config.DefaultsConfig,
	task *v1alpha1.Task,
	app *v1alpha1.App,
	space *v1alpha1.Space,
	containerCommand []string,
) (*tektonv1beta1.TaskRun, error) {
	userContainer, err := getUserContainer(task, app, space, containerCommand)
	step := tektonv1beta1.Step{}
	step.SetContainerFields(*userContainer)
	if err != nil {
		return nil, err
	}

	taskRun := &tektonv1beta1.TaskRun{
		ObjectMeta: makeObjectMeta(TaskRunName(task), task, app),
		Spec: tektonv1beta1.TaskRunSpec{
			TaskSpec: &tektonv1beta1.TaskSpec{
				Steps: []tektonv1beta1.Step{
					step,
				},
			},
			ServiceAccountName: app.Status.ServiceAccountName,
		},
	}

	if timeoutMins := cfg.TaskDefaultTimeoutMinutes; timeoutMins != nil {
		// https://tekton.dev/vault/pipelines-v0.19.0/taskruns/#configuring-the-failure-timeout

		// Values <= 0 mean infinite timeout.
		if *timeoutMins <= 0 { // Infinite timeout
			taskRun.Spec.Timeout = &metav1.Duration{Duration: 0}
		} else {
			taskRun.Spec.Timeout = &metav1.Duration{Duration: time.Duration(*timeoutMins) * time.Minute}
		}
	}

	if len(app.Status.Volumes) > 0 {
		userContainer := &taskRun.Spec.TaskSpec.Steps[0]
		// mapfs for volumes needs the extra permission.
		userContainer.SecurityContext = &corev1.SecurityContext{
			Privileged: ptr.Bool(true),
		}

		// Build NFS volume mounts.
		volumes, userVolumeMounts, fuseCommands, _, err := appresources.BuildVolumes(app.Status.Volumes)
		if err != nil {
			return nil, err
		}
		taskRun.Spec.TaskSpec.Volumes = append(taskRun.Spec.TaskSpec.Volumes, volumes...)
		userContainer.VolumeMounts = append(userContainer.VolumeMounts, userVolumeMounts...)

		// Grab the user's original command in a way that we can inject into a shell.
		desiredCommand := []string{}
		if len(userContainer.Command) > 0 {
			// Append to the existing array so we don't modify the userContainer.Command value.
			desiredCommand = append(desiredCommand, userContainer.Command...)
		} else {
			desiredCommand = append(desiredCommand, containerCommand...)
		}
		for _, arg := range userContainer.Args {
			desiredCommand = append(desiredCommand, shellescape.Quote(arg))
		}

		combinedStartCommand := append(fuseCommands, strings.Join(desiredCommand, " "))

		userContainer.Command = []string{"/bin/sh"}
		userContainer.Args = []string{"-c", strings.Join(combinedStartCommand, " ")}

		// XXX: Tekton doesn't allow hooking into the contianer's lifecycle so we rely on the underlying system to unmount.
	}

	if task.Spec.Terminated == true {
		taskRun.Spec.Status = tektonv1beta1.TaskRunSpecStatusCancelled
	}

	return taskRun, nil
}

func getUserContainer(
	task *v1alpha1.Task,
	app *v1alpha1.App,
	space *v1alpha1.Space,
	containerCommand []string,
) (*corev1.Container, error) {
	spec := app.Spec.Template.Spec.DeepCopy()

	// At this point in the lifecycle there should be exactly one container
	// spec in the App if the webhhook is working but create one to avoid
	// panics just in case.
	if len(spec.Containers) == 0 {
		spec.Containers = append(spec.Containers, corev1.Container{})
	}

	userContainer := &spec.Containers[0]
	userContainer.Name = v1alpha1.DefaultUserContainerName
	userContainer.Image = app.Status.Image

	// Inherit environment variables from App.
	containerEnv := []corev1.EnvVar{}
	containerEnv = append(containerEnv, space.Status.RuntimeConfig.Env...)
	containerEnv = append(containerEnv, userContainer.Env...)

	// Add in additinal CF style environment variables.
	containerEnv = append(containerEnv, appresources.BuildRuntimeEnvVars(appresources.CFTask, app)...)

	userContainer.Env = containerEnv

	// Explicitly disable stdin and tty allocation.
	userContainer.Stdin = false
	userContainer.TTY = false

	// userContainer.Command is the entrypoint to the App, it is set by the
	// --entrypoint flag during `kf push`, it is meant to override the
	// entrypoint in the container image. Threfore, only use the entrypoint
	// from the container image if no entrypoint is set on the App.
	if len(userContainer.Command) == 0 {
		userContainer.Command = containerCommand
	}

	// args from the image entrypoint is not overridable.
	overrideContainerArgs(userContainer, task)
	if err := overrideResourceRequests(userContainer, task); err != nil {
		return nil, err
	}
	// Task does not have probes.
	userContainer.ReadinessProbe = nil
	userContainer.LivenessProbe = nil
	userContainer.StartupProbe = nil

	return userContainer, nil
}

func overrideResourceRequests(container *corev1.Container, task *v1alpha1.Task) error {
	// Initialize resource requests to avoid panic.
	if container.Resources.Requests == nil {
		container.Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
	}
	requests := container.Resources.Requests

	if len(task.Spec.CPU) > 0 {
		cpuQuantity, err := resource.ParseQuantity(task.Spec.CPU)
		if err != nil {
			return err
		}
		requests[corev1.ResourceCPU] = cpuQuantity
	}

	if len(task.Spec.Memory) > 0 {
		memoryQuantity, err := resource.ParseQuantity(task.Spec.Memory)
		if err != nil {
			return err
		}
		requests[corev1.ResourceMemory] = memoryQuantity
	}

	if len(task.Spec.Disk) > 0 {
		diskQuantity, err := resource.ParseQuantity(task.Spec.Disk)
		if err != nil {
			return err
		}
		requests[corev1.ResourceEphemeralStorage] = diskQuantity
	}

	return nil
}

func overrideContainerArgs(container *corev1.Container, task *v1alpha1.Task) {
	// args precedence: Task command > App's container args (set by `--command` or `--args` flag in `kf push`).
	if len(task.Spec.Command) > 0 {
		container.Args = []string{task.Spec.Command}
	}
}
