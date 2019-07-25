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
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultMem     = resource.MustParse("1Gi")
	defaultStorage = resource.MustParse("1Gi")
	defaultCPU     = resource.MustParse("1")
)

// SetDefaults implements apis.Defaultable
func (k *App) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *AppSpec) SetDefaults(ctx context.Context) {
	// We require at least one container, so if there isn't one, set a blank
	// one.
	if len(k.Template.Spec.Containers) == 0 {
		k.Template.Spec.Containers = append(k.Template.Spec.Containers, corev1.Container{})
	}

	// Set default disk, RAM, and CPU limits on the application if they have not been custom set
	k.setResourceRequests(defaultMem, defaultStorage, defaultCPU)
}

func (k *AppSpec) setResourceRequests(memory resource.Quantity, storage resource.Quantity, cpu resource.Quantity) {
	userContainer := &k.Template.Spec.Containers[0]
	if userContainer.Resources.Requests == nil {
		userContainer.Resources.Requests = v1.ResourceList{}
	}

	if _, exists := userContainer.Resources.Requests[corev1.ResourceMemory]; !exists {
		userContainer.Resources.Requests[corev1.ResourceMemory] = memory
	}

	if _, exists := userContainer.Resources.Requests[corev1.ResourceEphemeralStorage]; !exists {
		userContainer.Resources.Requests[corev1.ResourceEphemeralStorage] = storage
	}

	if _, exists := userContainer.Resources.Requests[corev1.ResourceCPU]; !exists {
		userContainer.Resources.Requests[corev1.ResourceCPU] = cpu
	}
}
