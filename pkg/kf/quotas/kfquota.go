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
type KfQuota v1.ResourceList
type KfQuota2 = v1.ResourceList

// // GetName retrieves the name of the quota.
// func (k *KfQuota) GetName() string {
// 	return k.Name
// }

// // SetName sets the name of the quota.
// func (k *KfQuota) SetName(name string) {
// 	k.Name = name
// }

// GetMemory returns the quota for total memory in a space.
func (k KfQuota) GetMemory() (resource.Quantity, bool) {
	quantity, quotaExists := k[v1.ResourceMemory]
	return quantity, quotaExists
}

// SetMemory sets the quota for total memory in a space.
func (k KfQuota) SetMemory(memoryLimit resource.Quantity) {
	if k == nil {
		k = KfQuota{}
	}

	k[v1.ResourceMemory] = memoryLimit
}

// ResetMemory resets the quota for total memory in a space to unlimited.
func (k KfQuota) ResetMemory() {
	delete(k, v1.ResourceMemory)
}

// GetCPU returns the quota for total CPU in a space.
func (k KfQuota) GetCPU() (resource.Quantity, bool) {
	quantity, quotaExists := k[v1.ResourceCPU]
	return quantity, quotaExists
}

// SetCPU sets the quota for total CPU in a space.
func (k KfQuota) SetCPU(cpuLimit resource.Quantity) {
	if k == nil {
		k = KfQuota{}
	}

	k[v1.ResourceCPU] = cpuLimit
}

// ResetCPU resets the quota for total CPU in a space to unlimited.
func (k KfQuota) ResetCPU() {
	delete(k, v1.ResourceCPU)
}

// GetServices returns the quota for total number of routes in a space.
func (k KfQuota) GetServices() (resource.Quantity, bool) {
	quantity, quotaExists := k[v1.ResourceServices]
	return quantity, quotaExists
}

// SetServices sets the quota for total number of routes in a space.
func (k KfQuota) SetServices(numServices resource.Quantity) {
	if k == nil {
		k = KfQuota{}
	}

	k[v1.ResourceServices] = numServices
}

// ResetServices resets the quota for total number of routes in a space
// to unlimited.
func (k KfQuota) ResetServices() {
	delete(k, v1.ResourceServices)
}

// ToResourceList casts this alias back into a ResourceList.
func (k *KfQuota) ToResourceList() *v1.ResourceList {
	quotaList := v1.ResourceList(*k)
	return &quotaList
}

// NewFromResourceList casts a ResourceList into a KfQuota.
func NewFromResourceList(q *v1.ResourceList) *KfQuota {
	kfQuota := (*KfQuota)(q)
	return kfQuota
}

// NewKfQuota creates a new KfQuota.
func NewKfQuota() KfQuota {
	return KfQuota{}
}
