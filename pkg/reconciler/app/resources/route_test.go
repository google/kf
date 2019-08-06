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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestMakeRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		app    v1alpha1.App
		space  v1alpha1.Space
		assert func(t *testing.T, routes []v1alpha1.Route, claims []v1alpha1.RouteClaim)
	}{
		"configures correctly": {
			app: v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-name",
				},
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: "example.com", Path: "/some-path"},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, claims []v1alpha1.RouteClaim) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "route.Spec.AppName", "some-name", routes[0].Spec.AppName)
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
				testutil.AssertEqual(t, "route.Spec.Path", "/some-path", routes[0].Spec.Path)
			},
		},
		"creates claim route": {
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: "example.com", Path: "/some-path"},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, claims []v1alpha1.RouteClaim) {
				testutil.AssertEqual(t, "len(claims)", 1, len(claims))
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "example.com", "/some-path", ""),
					claims[0].Name,
				)
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", claims[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", claims[0].Spec.Hostname)
				testutil.AssertEqual(t, "route.Spec.Path", "/some-path", claims[0].Spec.Path)
			},
		},
		"no domain, uses space default": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com", Default: true},
							{Domain: "wrong.example.com", Default: false},
						},
					},
				},
			},
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: ""},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, claims []v1alpha1.RouteClaim) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
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
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: "some-domain", Path: "some-path"},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route, claims []v1alpha1.RouteClaim) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "some-domain", "some-path", "some-app-name"),
					routes[0].ObjectMeta.Name,
				)
				testutil.AssertEqual(t, "route.ObjectMeta.Labels", map[string]string{
					"a":                     "1",
					"b":                     "2",
					v1alpha1.NameLabel:      "some-app-name",
					v1alpha1.ManagedByLabel: "kf",
					v1alpha1.ComponentLabel: "route",
					v1alpha1.RouteHostname:  "some-hostname",
					v1alpha1.RouteDomain:    "some-domain",
					v1alpha1.RoutePath:      toBase36("/some-path"),
					v1alpha1.RouteAppName:   "some-app-name",
				}, routes[0].ObjectMeta.Labels)
				testutil.AssertEqual(
					t,
					"OwnerReferences",
					"some-app-name",
					routes[0].ObjectMeta.OwnerReferences[0].Name,
				)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			routes, claims, err := MakeRoutes(&tc.app, &tc.space)
			testutil.AssertNil(t, "err", err)
			tc.assert(t, routes, claims)
		})
	}
}

func ExampleMakeRouteLabels() {
	l := MakeRouteLabels(v1alpha1.RouteSpecFields{
		Hostname: "some-hostname",
		Domain:   "some-domain",
		Path:     "/some/path",
	})

	fmt.Println("Managed by:", l[v1alpha1.ManagedByLabel])
	fmt.Println("Component Label:", l[v1alpha1.ComponentLabel])
	fmt.Println("Route Hostname:", l[v1alpha1.RouteHostname])
	fmt.Println("Route Domain:", l[v1alpha1.RouteDomain])
	fmt.Println("Route Path (base-36):", l[v1alpha1.RoutePath])
	fmt.Printf("Number of Keys: %d\n", len(l))

	// Output: Managed by: kf
	// Component Label: route
	// Route Hostname: some-hostname
	// Route Domain: some-domain
	// Route Path (base-36): 2uusd3k2mp26d
	// Number of Keys: 5
}

func TestMakeRouteSelector(t *testing.T) {
	t.Parallel()

	s := MakeRouteSelector(v1alpha1.RouteSpecFields{
		Hostname: "some-host",
		Domain:   "some-domain",
		Path:     "some-path",
	})

	good := labels.Set{
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "route",
		v1alpha1.RouteHostname:  "some-host",
		v1alpha1.RouteDomain:    "some-domain",
		v1alpha1.RoutePath:      toBase36("/some-path"),
	}
	bad := labels.Set{
		v1alpha1.ManagedByLabel: "other-kf",
		v1alpha1.ComponentLabel: "other-route",
		v1alpha1.RouteHostname:  "some-other-host",
		v1alpha1.RouteDomain:    "some-other-host",
		v1alpha1.RoutePath:      toBase36("some-other-path"),
	}

	testutil.AssertEqual(t, "matches", true, s.Matches(good))
	testutil.AssertEqual(t, "doesn't match", false, s.Matches(bad))
}

func ExampleUnionMaps() {
	x := map[string]string{"a": "1", "b": "x", "c": "x"}
	y := map[string]string{"a": "1", "b": "2", "c": "3"}
	z := map[string]string{"a": "1", "b": "2", "c": "3"}

	result := UnionMaps(x, y, z)

	for _, key := range []string{"a", "b", "c"} {
		fmt.Printf("%s: %s\n", key, result[key])
	}

	// Output: a: 1
	// b: 2
	// c: 3
}
