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
	"sort"
)

// SetDefaults implements apis.Defaultable
func (k *Route) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpec) SetDefaults(ctx context.Context) {
	// Dedupe Service Names
	sort.Strings(k.KnativeServiceNames)
	var currentIdx int
	for i := 0; i < len(k.KnativeServiceNames); i++ {
		if i != 0 && k.KnativeServiceNames[i] == k.KnativeServiceNames[i-1] {
			continue
		}
		k.KnativeServiceNames[currentIdx] = k.KnativeServiceNames[i]
		currentIdx++
	}
	k.KnativeServiceNames = k.KnativeServiceNames[:currentIdx]
}
