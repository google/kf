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

// ToSpace casts this alias back into a Namespace.
func (k *KfSpace) ToSpace() *v1alpha1.Space {
	return (*v1alpha1.Space)(k)
}

// NewKfSpace creates a new KfSpace.
func NewKfSpace() KfSpace {
	return KfSpace{}
}
