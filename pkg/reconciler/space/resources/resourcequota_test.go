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

package resources

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ExampleMakeResourceQuota() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"
	mem, _ := resource.ParseQuantity("20Gi")
	cpu, _ := resource.ParseQuantity("800m")
	space.Spec.ResourceLimits.SpaceQuota = v1.ResourceList{
		v1.ResourceMemory: mem,
		v1.ResourceCPU:    cpu,
	}

	quota, err := MakeResourceQuota(space)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", ResourceQuotaName(space))
	fmt.Println("Managed by:", quota.Labels[managedByLabel])
	fmt.Println("Memory quota:", quota.Spec.Hard.Memory())
	fmt.Println("CPU quota:", quota.Spec.Hard.Cpu())

	// Output: Name: space-quota
	// Managed by: kf
	// Memory quota: 20Gi
	// CPU quota: 800m
}
