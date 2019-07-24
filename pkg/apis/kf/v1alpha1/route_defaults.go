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

	"github.com/google/kf/pkg/kf/algorithms"
)

// GenerateRouteName creates the deterministic name for a Route.
func GenerateRouteName(hostname, domain, path string) string {
	return GenerateName(hostname, domain, path)
}

// GenerateRouteNameFromSpec creates the deterministic name for a Route.
func GenerateRouteNameFromSpec(spec RouteSpecFields) string {
	return GenerateName(spec.Hostname, spec.Domain, spec.Path)
}

// SetDefaults implements apis.Defaultable
func (k *Route) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpec) SetDefaults(ctx context.Context) {
	k.AppNames = []string(algorithms.Dedupe(
		algorithms.Strings(k.AppNames),
	).(algorithms.Strings))
}

// SetSpaceDefaults sets the default values for the Route based on the space's
// settings.
func (k *Route) SetSpaceDefaults(space *Space) {
	k.Spec.SetSpaceDefaults(space)
}

// SetSpaceDefaults sets the default values for the RouteSpec based on the
// space's settings.
func (k *RouteSpec) SetSpaceDefaults(space *Space) {
	k.RouteSpecFields.SetSpaceDefaults(space)
}

// SetSpaceDefaults sets the default values for the RouteSpec based on the
// space's settings.
func (k *RouteSpecFields) SetSpaceDefaults(space *Space) {
	if k.Domain == "" {
		// Use space's default domain
		for _, domain := range space.Spec.Execution.Domains {
			if !domain.Default {
				continue
			}
			k.Domain = domain.Domain
			break
		}
	}
}
