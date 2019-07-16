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

package spaces

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// KfSpace provides a facade around v1alpha1.Space for accessing and mutating
// its values.
type KfSpace v1alpha1.Space

// GetName retrieves the name of the space.
func (k *KfSpace) GetName() string {
	return k.Name
}

// SetName sets the name of the space.
func (k *KfSpace) SetName(name string) {
	k.Name = name
}

// GetContainerRegistry gets the container registry for the space.
func (k *KfSpace) GetContainerRegistry() string {
	return k.Spec.BuildpackBuild.ContainerRegistry
}

// SetContainerRegistry sets the container registry for the space.
func (k *KfSpace) SetContainerRegistry(registry string) {
	k.Spec.BuildpackBuild.ContainerRegistry = registry
}

// GetQuota retrieves the space quota.
func (k *KfSpace) GetQuota() v1.ResourceList {
	return k.Spec.ResourceLimits.SpaceQuota
}

// DeleteQuota deletes the space quota.
func (k *KfSpace) DeleteQuota() error {
	k.Spec.ResourceLimits.SpaceQuota = nil
	return nil
}

// GetMemory returns the quota for total memory in a space.
func (k *KfSpace) GetMemory() (resource.Quantity, bool) {
	quantity, quotaExists := k.Spec.ResourceLimits.SpaceQuota[v1.ResourceMemory]
	return quantity, quotaExists
}

// SetMemory sets the quota for total memory in a space.
func (k *KfSpace) SetMemory(memoryLimit resource.Quantity) {
	if k.Spec.ResourceLimits.SpaceQuota == nil {
		k.Spec.ResourceLimits.SpaceQuota = v1.ResourceList{}
	}

	k.Spec.ResourceLimits.SpaceQuota[v1.ResourceMemory] = memoryLimit
}

// ResetMemory resets the quota for total memory in a space to unlimited.
func (k *KfSpace) ResetMemory() {
	delete(k.Spec.ResourceLimits.SpaceQuota, v1.ResourceMemory)
}

// GetCPU returns the quota for total CPU in a space.
func (k *KfSpace) GetCPU() (resource.Quantity, bool) {
	quantity, quotaExists := k.Spec.ResourceLimits.SpaceQuota[v1.ResourceCPU]
	return quantity, quotaExists
}

// SetCPU sets the quota for total CPU in a space.
func (k *KfSpace) SetCPU(cpuLimit resource.Quantity) {
	if k.Spec.ResourceLimits.SpaceQuota == nil {
		k.Spec.ResourceLimits.SpaceQuota = v1.ResourceList{}
	}

	k.Spec.ResourceLimits.SpaceQuota[v1.ResourceCPU] = cpuLimit
}

// ResetCPU resets the quota for total CPU in a space to unlimited.
func (k *KfSpace) ResetCPU() {
	delete(k.Spec.ResourceLimits.SpaceQuota, v1.ResourceCPU)
}

// GetServices returns the quota for total number of routes in a space.
func (k *KfSpace) GetServices() (resource.Quantity, bool) {
	quantity, quotaExists := k.Spec.ResourceLimits.SpaceQuota[v1.ResourceServices]
	return quantity, quotaExists
}

// SetServices sets the quota for total number of routes in a space.
func (k *KfSpace) SetServices(numServices resource.Quantity) {
	if k.Spec.ResourceLimits.SpaceQuota == nil {
		k.Spec.ResourceLimits.SpaceQuota = v1.ResourceList{}
	}

	k.Spec.ResourceLimits.SpaceQuota[v1.ResourceServices] = numServices
}

// ResetServices resets the quota for total number of routes in a space
// to unlimited.
func (k *KfSpace) ResetServices() {
	delete(k.Spec.ResourceLimits.SpaceQuota, v1.ResourceServices)
}

// ToSpace casts this alias back into a v1alpha.Space.
func (k *KfSpace) ToSpace() *v1alpha1.Space {
	return (*v1alpha1.Space)(k)
}

// NewFromSpace casts a v1alpha1.Space into a KfSpace.
func NewFromSpace(s *v1alpha1.Space) *KfSpace {
	kfSpace := (*KfSpace)(s)
	return kfSpace
}

// NewKfSpace creates a new KfSpace.
func NewKfSpace() KfSpace {
	return KfSpace{}
}
