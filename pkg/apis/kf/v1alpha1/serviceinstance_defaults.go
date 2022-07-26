// Copyright 2020 Google LLC
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
)

// SetDefaults implements apis.Defaultable.
func (instance *ServiceInstance) SetDefaults(ctx context.Context) {
	instance.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable.
func (spec *ServiceInstanceSpec) SetDefaults(ctx context.Context) {
	spec.ServiceType.SetDefaults(ctx)

	// Tags don't have defaults

	// ParametersFrom don't have defaults
}

// SetDefaults implements apis.Defaultable.
func (serviceType *ServiceType) SetDefaults(ctx context.Context) {
	if serviceType.UPS != nil {
		serviceType.UPS.SetDefaults(ctx)
	}

	if serviceType.Brokered != nil {
		serviceType.Brokered.SetDefaults(ctx)
	}

	if serviceType.OSB != nil {
		serviceType.OSB.SetDefaults(ctx)
	}
}

// SetDefaults implements apis.Defaultable.
func (instance *UPSInstance) SetDefaults(ctx context.Context) {
	// Nothing to default
}

// SetDefaults implements apis.Defaultable.
func (instance *BrokeredInstance) SetDefaults(ctx context.Context) {
	// Nothing to default
}

// SetDefaults implements apis.Defaultable.
func (instance *OSBInstance) SetDefaults(ctx context.Context) {
	if instance.ProgressDeadlineSeconds == 0 {
		instance.ProgressDeadlineSeconds = DefaultServiceInstanceProgressDeadlineSeconds
	}
}
