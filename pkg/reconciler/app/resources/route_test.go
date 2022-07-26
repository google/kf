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
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

func TestMakeRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		app    v1alpha1.App
		space  v1alpha1.Space
		assert func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding)
	}{
		"configures correctly": {
			app: v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-name",
				},
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight:          ptr.Int32(1),
							DestinationPort: ptr.Int32(9090),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(bindings)", 1, len(bindings))
				testutil.AssertEqual(t, "binding.Destination.ServiceName", "some-name", bindings[0].Destination.ServiceName)
				testutil.AssertEqual(t, "binding.Source.Domain", "example.com", bindings[0].Source.Domain)
				testutil.AssertEqual(t, "binding.Source.Hostname", "some-hostname", bindings[0].Source.Hostname)
				testutil.AssertEqual(t, "binding.Source.Path", "/some-path", bindings[0].Source.Path)
				testutil.AssertEqual(t, "binding.Destination.Port", int32(9090), bindings[0].Destination.Port)
				testutil.AssertEqual(t, "binding.Destination.Weight", int32(1), bindings[0].Destination.Weight)
			},
		},
		"creates claim route": {
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight: ptr.Int32(1),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "example.com", "/some-path"),
					routes[0].Name,
				)
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
				testutil.AssertEqual(t, "route.Spec.Path", "/some-path", routes[0].Spec.Path)
			},
		},
		"one claim per destination": {
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight: ptr.Int32(1),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
						{
							Weight:          ptr.Int32(1),
							DestinationPort: ptr.Int32(8080),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "len(bindings)", 2, len(bindings))
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "example.com", "/some-path"),
					routes[0].Name,
				)
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
				testutil.AssertEqual(t, "route.Spec.Path", "/some-path", routes[0].Spec.Path)
			},
		},
		"merges RouteWeightBindings": {
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
						{
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "len(bindings)", 1, len(bindings))
				testutil.AssertEqual(t, "binding.Destination.Weight", int32(2), bindings[0].Destination.Weight)
			},
		},
		"defaults destinationport": {
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(bindings)", 1, len(bindings))
				testutil.AssertEqual(t, "binding.Destination.Port", int32(v1alpha1.DefaultRouteDestinationPort), bindings[0].Destination.Port)
			},
		},
		"no domain, uses space default": {
			space: v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					NetworkConfig: v1alpha1.SpaceStatusNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
							{Domain: "wrong.example.com"},
						},
					},
				},
			},
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight: ptr.Int32(1),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(bindings)", 1, len(bindings))
				testutil.AssertEqual(t, "binding.Source.Domain", "example.com", bindings[0].Source.Domain)
				testutil.AssertEqual(t, "binding.Source.Hostname", "some-hostname", bindings[0].Source.Hostname)
			},
		},
		"merges default domain with existing": {
			space: v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					NetworkConfig: v1alpha1.SpaceStatusNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
						},
					},
				},
			},
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight: ptr.Int32(2),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "",
							},
						},
						{
							Weight: ptr.Int32(3),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(bindings)", 1, len(bindings))
				testutil.AssertEqual(t, "binding.Source.Domain", "example.com", bindings[0].Source.Domain)
				testutil.AssertEqual(t, "binding.Source.Hostname", "some-hostname", bindings[0].Source.Hostname)
				testutil.AssertEqual(t, "binding.Destination.Weight", int32(5), bindings[0].Destination.Weight)
			},
		},
		"ObjectMeta": {
			space: v1alpha1.Space{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-space-name",
				},
			},
			app: v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-app-name",
					Labels:    map[string]string{"a": "1", "b": "2"},
				},
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteWeightBinding{
						{
							Weight:          ptr.Int32(1),
							DestinationPort: ptr.Int32(9999),
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "some-domain",
								Path:     "some-path",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, bindings []v1alpha1.QualifiedRouteBinding) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "some-domain", "some-path"),
					routes[0].ObjectMeta.Name,
				)
				testutil.AssertEqual(t, "route.ObjectMeta.Labels", map[string]string{
					v1alpha1.ManagedByLabel: "kf",
					v1alpha1.ComponentLabel: "route",
				}, routes[0].ObjectMeta.Labels)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			routes, bindings, err := MakeRoutes(&tc.app, &tc.space)
			testutil.AssertNil(t, "err", err)
			tc.assert(t, routes, bindings)
		})
	}
}

func ExampleMakeRouteLabels() {
	l := MakeRouteLabels()

	fmt.Println("Managed by:", l[v1alpha1.ManagedByLabel])
	fmt.Println("Component Label:", l[v1alpha1.ComponentLabel])
	fmt.Printf("Number of Keys: %d\n", len(l))

	// Output: Managed by: kf
	// Component Label: route
	// Number of Keys: 2
}
