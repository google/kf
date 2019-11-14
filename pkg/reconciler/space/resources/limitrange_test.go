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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ExampleMakeLimitRange() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"
	mem, _ := resource.ParseQuantity("1Gi")
	cpu, _ := resource.ParseQuantity("100m")
	defaultRequestLimit := v1.ResourceList{
		v1.ResourceMemory: mem,
		v1.ResourceCPU:    cpu,
	}
	limit := v1.LimitRangeItem{
		Type:           v1.LimitTypePod,
		DefaultRequest: defaultRequestLimit,
	}
	space.Spec.ResourceLimits.ResourceDefaults = []v1.LimitRangeItem{limit}

	limitRange, err := MakeLimitRange(space)
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Println("Name:", LimitRangeName(space)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Managed by:", limitRange.Labels[managedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Limit type:", limitRange.Spec.Limits[0].Type); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Default memory request:", limitRange.Spec.Limits[0].DefaultRequest.Memory()); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Default cpu request:", limitRange.Spec.Limits[0].DefaultRequest.Cpu()); err != nil {
		panic(err)
	}

	// Output: Name: space-limit-range
	// Managed by: kf
	// Limit type: Pod
	// Default memory request: 1Gi
	// Default cpu request: 100m
}
