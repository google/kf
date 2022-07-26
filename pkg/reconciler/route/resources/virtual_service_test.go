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
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeRouteSpecFields(host, domain, path string) v1alpha1.RouteSpecFields {
	return v1alpha1.RouteSpecFields{
		Hostname: host,
		Domain:   domain,
		Path:     path,
	}
}

func makeRouteSpecFieldsStr(host, domain, path string) string {
	return v1alpha1.RouteSpecFields{
		Hostname: host,
		Domain:   domain,
		Path:     path,
	}.String()
}

func makeRouteBinding(host, domain, path, appName string, weight int32) v1alpha1.QualifiedRouteBinding {
	return makeRouteBindingWithPort(host, domain, path, appName, weight, v1alpha1.DefaultRouteDestinationPort)
}

func makeRouteBindingWithPort(host, domain, path, appName string, weight, port int32) v1alpha1.QualifiedRouteBinding {
	return v1alpha1.QualifiedRouteBinding{
		Source: makeRouteSpecFields(host, domain, path),
		Destination: v1alpha1.RouteDestination{
			Weight:      weight,
			ServiceName: appName,
			Port:        port,
		},
	}
}

func makeAppDestination(appName string, weight int32) v1alpha1.RouteDestination {
	return makeAppDestinationWithPort(appName, weight, v1alpha1.DefaultRouteDestinationPort)
}

func makeAppDestinationWithPort(appName string, weight, port int32) v1alpha1.RouteDestination {
	return v1alpha1.RouteDestination{
		Weight:      weight,
		ServiceName: appName,
		Port:        port,
	}
}

func makeRouteServiceDestination(name, scheme, host, path string) v1alpha1.RouteServiceDestination {
	return v1alpha1.RouteServiceDestination{
		Name: name,
		RouteServiceURL: &v1alpha1.RouteServiceURL{
			Scheme: scheme,
			Host:   host,
			Path:   path,
		},
	}
}

func makeRouteServiceDestinationWithPort(name, scheme, host, path string, port int32) v1alpha1.RouteServiceDestination {
	rsDestination := makeRouteServiceDestination(name, scheme, host, path)
	hostWithPort := fmt.Sprintf("%s:%d", rsDestination.RouteServiceURL.Host, port)
	rsDestination.RouteServiceURL.Host = hostWithPort
	return rsDestination
}

func makeRoute(host, domain, path, namespace string) *v1alpha1.Route {
	return &v1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			// The names here aren't real or tied to any particular implementation
			// they exist to ensure the name gets propagated correctly in the owner
			// references on the VirtualService.
			Name:      v1alpha1.GenerateName("fake-route", host, domain, path),
			Namespace: namespace,
		},
		Spec: v1alpha1.RouteSpec{
			RouteSpecFields: makeRouteSpecFields(host, domain, path),
		},
	}
}

func TestMakeVirtualService(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Routes               []*v1alpha1.Route
		Bindings             map[string]RouteBindingSlice
		RouteServiceBindings map[string][]v1alpha1.RouteServiceDestination
		SpaceDomain          v1alpha1.SpaceDomain
		assertErr            error
	}{
		"empty list of routes": {
			assertErr: errors.New("routes must not be empty"),
		},
		"single route": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"blank host": {
			Routes: []*v1alpha1.Route{
				makeRoute("", "example.com", "/some-path", "some-namespace"),
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"single app binding": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("some-app", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"wildcard domain": {
			Routes: []*v1alpha1.Route{
				makeRoute("*", "example.com", "", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("*", "example.com", ""): []v1alpha1.RouteDestination{
					makeAppDestination("app-1", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"wildcard route works as backup": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
				makeRoute("*", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("app-1", 1),
				},
				makeRouteSpecFieldsStr("*", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("backup-app", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"multiple apps per route with different weights": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("app-1", 2), makeAppDestination("app-2", 1), makeAppDestination("app-3", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"longest path first": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com/", "", "some-namespace"),
				makeRoute("some-host", "example.com/", "/foo", "some-namespace"),
				makeRoute("some-host", "example.com/", "/foo/bar", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com/", ""): []v1alpha1.RouteDestination{
					makeAppDestination("should-be-third", 1),
				},
				makeRouteSpecFieldsStr("some-host", "example.com/", "/foo"): []v1alpha1.RouteDestination{
					makeAppDestination("should-be-second", 1),
				},
				makeRouteSpecFieldsStr("some-host", "example.com/", "/foo/bar"): []v1alpha1.RouteDestination{
					makeAppDestination("should-be-first", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"destination ports": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com/", "", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com/", ""): []v1alpha1.RouteDestination{
					makeAppDestinationWithPort("myapp", 1, 8080), makeAppDestinationWithPort("myapp", 1, 9999),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"custom gateway": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("some-app", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"route service": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "", "some-namespace"),
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("some-app", 1),
				},
			},
			RouteServiceBindings: map[string][]v1alpha1.RouteServiceDestination{
				makeRouteSpecFieldsStr("some-host", "example.com", ""): {
					makeRouteServiceDestination("some-route-svc", "http", "some-route-service.com", ""),
				},
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): {
					makeRouteServiceDestinationWithPort("another-route-svc", "https", "another-route-service.com", "/fake-path", 443),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/some-gateway",
			},
		},
		"internal routing": {
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path", "some-namespace"),
			},
			Bindings: map[string]RouteBindingSlice{
				makeRouteSpecFieldsStr("some-host", "example.com", "/some-path"): []v1alpha1.RouteDestination{
					makeAppDestination("some-app", 1),
				},
			},
			SpaceDomain: v1alpha1.SpaceDomain{
				Domain:      "example.com",
				GatewayName: "kf/internal-gateway",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			actualVS, actualErr := MakeVirtualService(tc.Routes, tc.Bindings, tc.RouteServiceBindings, &tc.SpaceDomain)
			testutil.AssertErrorsEqual(t, tc.assertErr, actualErr)
			testutil.AssertGoldenJSONContext(t, "virtualservice", actualVS, map[string]interface{}{
				"routes":               tc.Routes,
				"routeBindings":        convertBindingsForContext(tc.Routes, tc.Bindings),
				"routeServiceBindings": convertRouteServiceBindingsForContext(tc.Routes, tc.RouteServiceBindings),
				"spaceDomain":          tc.SpaceDomain,
			})

			// If the VS already passed the above tests, check against those structs
			// as golden to ensure it's valid. Don't check against the .golden files
			// because reading/updating them N times will get expensive.
			//if tc.assertErr == nil {
			//	for i := 0; i < 20; i++ {
			//		t.Run(fmt.Sprintf("iteration %d", i), func(t *testing.T) {
			//			iterVS, iterErr := MakeVirtualService(tc.Routes, tc.Bindings, tc.RouteServiceBindings, &tc.SpaceDomain)
			//			testutil.AssertErrorsEqual(t, tc.assertErr, iterErr)
			//			testutil.AssertEqual(t, "virtualService", actualVS, iterVS)
			//		})
			//	}
			//}
		})
	}
}

// Convert map of RouteSpecFields to []RouteDestination back to a list of QualifiedRouteBindings.
// This is only used to support the JSON marshaling for the test context output.
func convertBindingsForContext(routes []*v1alpha1.Route, appBindingsMap map[string]RouteBindingSlice) []*v1alpha1.QualifiedRouteBinding {
	var appBindings []*v1alpha1.QualifiedRouteBinding
	for _, route := range routes {
		rsf := route.Spec.RouteSpecFields
		appDestinations := appBindingsMap[rsf.String()]
		for _, destination := range appDestinations {
			appBinding := &v1alpha1.QualifiedRouteBinding{
				Source:      rsf,
				Destination: destination,
			}
			appBindings = append(appBindings, appBinding)
		}
	}
	return appBindings
}

// Convert map of RouteSpecFields to []RouteServiceDestination to a list of RouteServiceBindings.
// This is only used to support the JSON marshaling for the test context output.
func convertRouteServiceBindingsForContext(routes []*v1alpha1.Route, routeServiceBindingsMap map[string][]v1alpha1.RouteServiceDestination) []v1alpha1.RouteServiceBinding {
	var routeServiceBindings []v1alpha1.RouteServiceBinding
	for _, route := range routes {
		rsf := route.Spec.RouteSpecFields
		routeServiceDestinations := routeServiceBindingsMap[rsf.String()]
		if len(routeServiceDestinations) > 0 {
			routeServiceBinding := v1alpha1.RouteServiceBinding{
				Source:      rsf,
				Destination: routeServiceDestinations[len(routeServiceDestinations)-1].RouteServiceURL,
			}
			routeServiceBindings = append(routeServiceBindings, routeServiceBinding)
		}
	}
	return routeServiceBindings
}

func Test_normalizeRouteWeights(t *testing.T) {
	cases := map[string]struct {
		weights  RouteBindingSlice
		expected RouteBindingSlice
	}{
		"0 weights": {
			weights:  RouteBindingSlice{},
			expected: RouteBindingSlice{},
		},
		"1 weight": {
			weights: RouteBindingSlice{
				{ServiceName: "a", Weight: 1000},
			},
			expected: RouteBindingSlice{
				{ServiceName: "a", Weight: 100},
			},
		},
		"2 even splits": {
			weights: RouteBindingSlice{
				{ServiceName: "a", Weight: 100},
				{ServiceName: "b", Weight: 100},
			},
			expected: RouteBindingSlice{
				{ServiceName: "a", Weight: 50},
				{ServiceName: "b", Weight: 50},
			},
		},
		"3 even splits": {
			weights: RouteBindingSlice{
				{ServiceName: "a", Weight: 100},
				{ServiceName: "b", Weight: 100},
				{ServiceName: "c", Weight: 100},
			},
			expected: RouteBindingSlice{
				{ServiceName: "a", Weight: 34},
				{ServiceName: "b", Weight: 33},
				{ServiceName: "c", Weight: 33},
			},
		},
		"uneven split": {
			weights: RouteBindingSlice{
				{ServiceName: "a", Weight: 2},
				{ServiceName: "b", Weight: 1},
				{ServiceName: "c", Weight: 1},
			},
			expected: RouteBindingSlice{
				{ServiceName: "a", Weight: 50},
				{ServiceName: "b", Weight: 25},
				{ServiceName: "c", Weight: 25},
			},
		},
		"uneven split same app": {
			weights: RouteBindingSlice{
				{ServiceName: "a", Port: 8080, Weight: 3},
				{ServiceName: "a", Port: 80, Weight: 1},
			},
			expected: RouteBindingSlice{
				{ServiceName: "a", Port: 80, Weight: 25},
				{ServiceName: "a", Port: 8080, Weight: 75},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// Check that it's deterministic by running 100 times
			for i := 0; i < 100; i++ {
				t.Run(fmt.Sprintf("iteration %d", i), func(t *testing.T) {
					actual := normalizeRouteWeights(tc.weights)
					testutil.AssertEqual(t, "weights", tc.expected, actual)
				})
			}
		})
	}
}
