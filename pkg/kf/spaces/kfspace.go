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

import v1 "k8s.io/api/core/v1"

// KfSpace provides a facade around v1.Namespace for accessing and mutating its
// values.
type KfSpace v1.Namespace

// GetName retrieves the name of the space.
func (k *KfSpace) GetName() string {
	return k.Name
}

// SetName sets the name of the space.
func (k *KfSpace) SetName(name string) {
	k.Name = name
}

// ToNamespace casts this alias back into a Namespace.
func (k *KfSpace) ToNamespace() *v1.Namespace {
	ns := v1.Namespace(*k)
	return &ns
}

// NewKfSpace creates a new KfSpace.
func NewKfSpace() KfSpace {
	return KfSpace{}
}
