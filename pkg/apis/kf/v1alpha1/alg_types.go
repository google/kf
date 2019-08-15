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

	"github.com/google/kf/pkg/kf/algorithms"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/v1alpha3"
)

// TODO(poy): This file SHOULD be generated, but it has been written by hand.

// OwnerReferences implements the necessary interfaces for the algorithms
// package.
type OwnerReferences []metav1.OwnerReference

// Set implements Interface.
func (d OwnerReferences) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(OwnerReferences)[i] = b.(OwnerReferences)[j]
}

// Append implements Interface.
func (d OwnerReferences) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(OwnerReferences)...)
}

// Clone implements Interface.
func (d OwnerReferences) Clone() algorithms.Interface {
	return append(OwnerReferences{}, d...)
}

// Slice implements Interface.
func (d OwnerReferences) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d OwnerReferences) Len() int {
	return len(d)
}

// Less implements Interface.
func (d OwnerReferences) Less(i int, j int) bool {
	return d[i].UID < d[j].UID
}

// Swap implements Interface.
func (d OwnerReferences) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
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
	// We don't want to lose default information.
	if d[i].Domain == d[j].Domain {
		d[i].Default = d[i].Default || d[j].Default
		d[j].Default = d[i].Default || d[j].Default
	}

	return d[i].Domain < d[j].Domain
}

// Swap implements Interface.
func (d SpaceDomains) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

// Routes implements the necessary interfaces for the algorithms package.
type ServiceBindings []servicecatalogv1beta1.ServiceBinding

// Set implements Interface.
func (d ServiceBindings) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(ServiceBindings)[i] = b.(ServiceBindings)[j]
}

// Append implements Interface.
func (d ServiceBindings) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(ServiceBindings)...)
}

// Clone implements Interface.
func (s ServiceBindings) Clone() algorithms.Interface {
	return append(ServiceBindings{}, s...)
}

// Slice implements Interface.
func (s ServiceBindings) Slice(i int, j int) algorithms.Interface {
	return s[i:j]
}

// Len implements Interface.
func (s ServiceBindings) Len() int {
	return len(s)
}

// Less implements Interface.
func (s ServiceBindings) Less(i int, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap implements Interface.
func (s ServiceBindings) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

// Routes implements the necessary interfaces for the algorithms package.
type Routes []Route

// Set implements Interface.
func (d Routes) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(Routes)[i] = b.(Routes)[j]
}

// Append implements Interface.
func (d Routes) Append(a algorithms.Interface) algorithms.Interface {
	return append(d, a.(Routes)...)
}

// Clone implements Interface.
func (d Routes) Clone() algorithms.Interface {
	return append(Routes{}, d...)
}

// Slice implements Interface.
func (d Routes) Slice(i int, j int) algorithms.Interface {
	return d[i:j]
}

// Len implements Interface.
func (d Routes) Len() int {
	return len(d)
}

// Less implements Interface.
func (d Routes) Less(i int, j int) bool {
	return d[i].Name < d[j].Name
}

// Swap implements Interface.
func (d Routes) Swap(i int, j int) {
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
func (d RouteSpecFieldsSlice) Less(i int, j int) bool {
	// TODO(https://github.com/knative/pkg/issues/542):
	// We can't garuntee that the path will have the '/' or not
	// because webhooks can't yet modify slices.
	d[i].Path = path.Join("/", d[i].Path)
	d[j].Path = path.Join("/", d[j].Path)

	return GenerateRouteNameFromSpec(d[i], "") < GenerateRouteNameFromSpec(d[j], "")
}

// Swap implements Interface.
func (d RouteSpecFieldsSlice) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}
