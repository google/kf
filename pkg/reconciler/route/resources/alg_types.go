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
	"github.com/google/kf/pkg/kf/algorithms"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/v1alpha3"
)

// TODO(poy): This file SHOULD be generated, but it has been written by hand.

// OwnerReferences implements the necessary interfaces for the algorithms
// package.
type OwnerReferences []metav1.OwnerReference

// Set implements Interface.
func (o OwnerReferences) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(OwnerReferences)[i] = b.(OwnerReferences)[j]
}

// Append implements Interface.
func (o OwnerReferences) Append(a algorithms.Interface) algorithms.Interface {
	return append(o, a.(OwnerReferences)...)
}

// Clone implements Interface.
func (o OwnerReferences) Clone() algorithms.Interface {
	return append(OwnerReferences{}, o...)
}

// Slice implements Interface.
func (o OwnerReferences) Slice(i int, j int) algorithms.Interface {
	return o[i:j]
}

// Len implements Interface.
func (o OwnerReferences) Len() int {
	return len(o)
}

// Less implements Interface.
func (o OwnerReferences) Less(i int, j int) bool {
	return o[i].UID < o[j].UID
}

// Swap implements Interface.
func (o OwnerReferences) Swap(i int, j int) {
	o[i], o[j] = o[j], o[i]
}

// HTTPRoutes implements the necessary interfaces for the algorithms
// package.
type HTTPRoutes []v1alpha3.HTTPRoute

// Set implements Interface.
func (h HTTPRoutes) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(HTTPRoutes)[i] = b.(HTTPRoutes)[j]
}

// Append implements Interface.
func (h HTTPRoutes) Append(a algorithms.Interface) algorithms.Interface {
	return append(h, a.(HTTPRoutes)...)
}

// Clone implements Interface.
func (h HTTPRoutes) Clone() algorithms.Interface {
	return append(HTTPRoutes{}, h...)
}

// Slice implements Interface.
func (h HTTPRoutes) Slice(i int, j int) algorithms.Interface {
	return h[i:j]
}

// Len implements Interface.
func (h HTTPRoutes) Len() int {
	return len(h)
}

// Less implements Interface.
func (h HTTPRoutes) Less(i int, j int) bool {
	f := func(h v1alpha3.HTTPRoute) string {
		var m string
		for _, s := range h.Match {
			if s.URI == nil {
				continue
			}
			m += s.URI.Exact + s.URI.Prefix + s.URI.Suffix + s.URI.Regex
		}
		return m
	}

	return f(h[i]) < f(h[j])
}

// Swap implements Interface.
func (h HTTPRoutes) Swap(i int, j int) {
	h[i], h[j] = h[j], h[i]
}
