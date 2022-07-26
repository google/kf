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

package algorithms

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// Subjects implements Interface for a slice of Kubernetes RBAC Subjects.
type Subjects []rbacv1.Subject

var _ Interface = (Subjects)(nil)

// Len implements Interface.
func (s Subjects) Len() int {
	return len(s)
}

// Swap implements Interface.
func (s Subjects) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements Interface.
func (s Subjects) Less(i, j int) bool {
	lhs, rhs := s[i], s[j]

	if lhs.Kind != rhs.Kind {
		return lhs.Kind < rhs.Kind
	}

	if lhs.Name != rhs.Name {
		return lhs.Name < rhs.Name
	}

	return lhs.Namespace < rhs.Namespace
}

// Set implements Interface.
func (s Subjects) Set(i int, a Interface, j int, b Interface) {
	a.(Subjects)[i] = b.(Subjects)[j]
}

// Clone implements Interface.
func (s Subjects) Clone() Interface {
	return append(Subjects{}, s...)
}

// Append implements Interface.
func (s Subjects) Append(a Interface) Interface {
	return append(s, a.(Subjects)...)
}

// Slice implements Interface.
func (s Subjects) Slice(i, j int) Interface {
	return s[i:j]
}

// Contains checks if a Subject (Name + Kind) is contained in the Subjects slice.
func (s Subjects) Contains(name, kind string) (bool, int) {
	var index = -1
	for i, subject := range s {
		if subject.Name == name && subject.Kind == kind {
			index = i
		}
	}
	if index > -1 {
		return true, index
	}
	return false, -1
}
