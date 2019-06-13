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

package quotas

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// KfQuota provides a facade around v1.ResourceQuota for accessing and mutating its
// values.
type KfQuota v1.ResourceQuota

// GetName retrieves the name of the quota.
func (k *KfQuota) GetName() string {
	return k.Name
}

// SetName sets the name of the quota.
func (k *KfQuota) SetName(name string) {
	k.Name = name
}

// TODO: set and get actual resource quota fields
func (k *KfQuota) GetMemory() resource.Quantity {
	return k.Spec.Hard[v1.ResourceMemory]
}

func (k *KfQuota) SetMemory(memoryLimit resource.Quantity) {
	if k.Spec.Hard == nil {
		k.Spec.Hard = v1.ResourceList{}
	}

	k.Spec.Hard[v1.ResourceMemory] = memoryLimit
}

func (k *KfQuota) GetCPU() resource.Quantity {
	return k.Spec.Hard[v1.ResourceCPU]
}

func (k *KfQuota) SetCPU(cpuLimit resource.Quantity) {
	if k.Spec.Hard == nil {
		k.Spec.Hard = v1.ResourceList{}
	}

	k.Spec.Hard[v1.ResourceCPU] = cpuLimit
}

func (k *KfQuota) GetServices() resource.Quantity {
	return k.Spec.Hard[v1.ResourceServices]
}

func (k *KfQuota) SetServices(numServices resource.Quantity) {
	if k.Spec.Hard == nil {
		k.Spec.Hard = v1.ResourceList{}
	}

	k.Spec.Hard[v1.ResourceServices] = numServices
}

// ToResourceQuota casts this alias back into a ResourceQuota.
func (k *KfQuota) ToResourceQuota() *v1.ResourceQuota {
	quota := v1.ResourceQuota(*k)
	return &quota
}

// NewFromResourceQuota casts a ResourceQuota into a KfQuota.
func NewFromResourceQuota(q *v1.ResourceQuota) *KfQuota {
	kfQuota := (*KfQuota)(q)
	return kfQuota
}

// NewKfQuota creates a new KfQuota.
func NewKfQuota() KfQuota {
	return KfQuota{}
}
