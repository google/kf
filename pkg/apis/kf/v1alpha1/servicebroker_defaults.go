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

package v1alpha1

import "context"

// SetDefaults implements apis.Defaultable.
func (sb *ServiceBroker) SetDefaults(ctx context.Context) {
	sb.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable.
func (spec *ServiceBrokerSpec) SetDefaults(ctx context.Context) {
	spec.CommonServiceBrokerSpec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable.
func (sb *ClusterServiceBroker) SetDefaults(ctx context.Context) {
	sb.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable.
func (spec *ClusterServiceBrokerSpec) SetDefaults(ctx context.Context) {
	spec.CommonServiceBrokerSpec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable.
func (spec *CommonServiceBrokerSpec) SetDefaults(ctx context.Context) {
	if spec.UpdateRequests == 0 {
		spec.UpdateRequests = 1
	}
}
