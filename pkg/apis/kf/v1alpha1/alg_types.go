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
	"path"
	"strings"

	"github.com/google/kf/v2/pkg/kf/algorithms"
	corev1 "k8s.io/api/core/v1"
)

// TODO(poy): This file SHOULD be generated, but it has been written by hand.

// SpaceDomains implements the necessary interfaces for the algorithms
// package.
type SpaceDomains []SpaceDomain

// Set implements Interface.
func (d SpaceDomains) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(SpaceDomains)[i] = b.(SpaceDomains)[j]
}

// Append implements Interface.
func (d SpaceDomains) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(SpaceDomains)...)
}

// Clone implements Interface.
func (d SpaceDomains) Clone() algorithms.Interface {
	return append(SpaceDomains{}, d...)
}

// Slice implements Interface.
func (d SpaceDomains) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d SpaceDomains) Len() int {
	return len(d)
}

// Less implements Interface.
func (d SpaceDomains) Less(i int, j int) bool {
	return d[i].Domain < d[j].Domain
}

// Swap implements Interface.
func (d SpaceDomains) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

// RouteSpecFieldsSlice implements the necessary interfaces for the algorithms
// package.
type RouteSpecFieldsSlice []RouteSpecFields

// Set implements Interface.
func (d RouteSpecFieldsSlice) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(RouteSpecFieldsSlice)[i] = b.(RouteSpecFieldsSlice)[j]
}

// Append implements Interface.
func (d RouteSpecFieldsSlice) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(RouteSpecFieldsSlice)...)
}

// Clone implements Interface.
func (d RouteSpecFieldsSlice) Clone() algorithms.Interface {
	return append(RouteSpecFieldsSlice{}, d...)
}

// Slice implements Interface.
func (d RouteSpecFieldsSlice) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d RouteSpecFieldsSlice) Len() int {
	return len(d)
}

// Less implements Interface.
// This is used to sort RouteSpecFields in an order that makes sense for generating VirtualService rules.
// RouteSpecFields are sorted alphabetically by hostname and domain (though the domain should be the same for all RSFs being compared).
// RSFs with the "*" host are listed last, since they are the most general.
// Within RSFs with the same hostname + domain, the ones with longer paths come first, since they are more specific and
// should be evaluated first in the VS.
func (d RouteSpecFieldsSlice) Less(i int, j int) bool {
	// TODO(https://github.com/knative/pkg/issues/542):
	// We can't garuntee that the path will have the '/' or not
	// because webhooks can't yet modify slices.
	d[i].Path = path.Join("/", d[i].Path)
	d[j].Path = path.Join("/", d[j].Path)

	if d[i].Hostname != d[j].Hostname {
		if strings.HasPrefix(d[i].Hostname, "*") {
			return false
		}
		if strings.HasPrefix(d[j].Hostname, "*") {
			return true
		}
		return d[i].Hostname < d[j].Hostname
	}

	// RouteSpecFields sort is only used by the VirtualService reconciler currently,
	// which uses a list of RSFs with the same domain, so this case shouldn't happen.
	if d[i].Domain != d[j].Domain {
		return d[i].Domain < d[j].Domain
	}

	return len(d[i].Path) > len(d[j].Path)
}

// Swap implements Interface.
func (d RouteSpecFieldsSlice) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

// ObjectReferences implements the necessary interfaces for the algorithms
// package.
type ObjectReferences []corev1.ObjectReference

// Set implements Interface.
func (d ObjectReferences) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(ObjectReferences)[i] = b.(ObjectReferences)[j]
}

// Append implements Interface.
func (d ObjectReferences) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(ObjectReferences)...)
}

// Clone implements Interface.
func (d ObjectReferences) Clone() algorithms.Interface {
	return append(ObjectReferences{}, d...)
}

// Slice implements Interface.
func (d ObjectReferences) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d ObjectReferences) Len() int {
	return len(d)
}

// Less implements Interface.
func (d ObjectReferences) Less(i int, j int) bool {
	return d[i].Name < d[j].Name
}

// Swap implements Interface.
func (d ObjectReferences) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

// LocalObjectReferences implements the necessary interfaces for the algorithms
// package.
type LocalObjectReferences []corev1.LocalObjectReference

// Set implements Interface.
func (d LocalObjectReferences) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(LocalObjectReferences)[i] = b.(LocalObjectReferences)[j]
}

// Append implements Interface.
func (d LocalObjectReferences) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(LocalObjectReferences)...)
}

// Clone implements Interface.
func (d LocalObjectReferences) Clone() algorithms.Interface {
	return append(LocalObjectReferences{}, d...)
}

// Slice implements Interface.
func (d LocalObjectReferences) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d LocalObjectReferences) Len() int {
	return len(d)
}

// Less implements Interface.
func (d LocalObjectReferences) Less(i int, j int) bool {
	return d[i].Name < d[j].Name
}

// Swap implements Interface.
func (d LocalObjectReferences) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}
