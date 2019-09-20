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

package resources_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"github.com/knative/serving/pkg/network"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
)

func makeRouteSpecFields(host, domain, path string) v1alpha1.RouteSpecFields {
	return v1alpha1.RouteSpecFields{
		Hostname: host,
		Domain:   domain,
		Path:     path,
	}
}

func makeRoute(host, domain, path string) *v1alpha1.Route {
	return &v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			RouteSpecFields: makeRouteSpecFields(host, domain, path),
		},
	}
}

func makeRouteClaim(host, domain, path string) *v1alpha1.RouteClaim {
	return &v1alpha1.RouteClaim{
		Spec: v1alpha1.RouteClaimSpec{
			RouteSpecFields: makeRouteSpecFields(host, domain, path),
		},
	}
}

func TestMakeVirtualService(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Claims []*v1alpha1.RouteClaim
		Routes []*v1alpha1.Route
		Assert func(t *testing.T, v *networking.VirtualService, err error)
	}{
		"empty list of claims": {
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertErrorsEqual(t, errors.New("claims must not be empty"), err)
			},
		},
		"proper Meta": {
			Claims: []*v1alpha1.RouteClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some-namespace",
						Labels:    map[string]string{"a": "1", "b": "2"},
					},
					Spec: v1alpha1.RouteClaimSpec{
						RouteSpecFields: makeRouteSpecFields("some-host", "example.com", "some-path"),
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "TypeMeta", metav1.TypeMeta{
					APIVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
				}, v.TypeMeta)

				route := makeRoute("some-host", "example.com", "/some-path")

				testutil.AssertEqual(t, "ObjectMeta", metav1.ObjectMeta{
					Name:      v1alpha1.GenerateName(route.Spec.Hostname, route.Spec.Domain),
					Namespace: v1alpha1.KfNamespace,
					Labels: map[string]string{
						resources.ManagedByLabel: "kf",
						v1alpha1.ComponentLabel:  "virtualservice",
						v1alpha1.RouteHostname:   "some-host",
						v1alpha1.RouteDomain:     "example.com",
					},
					Annotations: map[string]string{
						"domain":   "example.com",
						"hostname": "some-host",
						"space":    "some-namespace",
					},
				}, v.ObjectMeta)
			},
		},
		"Path Matchers": {
			Claims: []*v1alpha1.RouteClaim{
				{
					Spec: v1alpha1.RouteClaimSpec{
						RouteSpecFields: makeRouteSpecFields("some-host", "example.com", "/some-path"),
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 1, len(v.Spec.HTTP))
				testutil.AssertEqual(t, "HTTP Match len", 1, len(v.Spec.HTTP[0].Match))
				testutil.AssertEqual(t, "HTTP Match", networking.HTTPMatchRequest{
					URI: &istio.StringMatch{
						Regex: "^/some-path(/.*)?",
					},
				}, v.Spec.HTTP[0].Match[0])
			},
		},
		"Route": {
			Claims: []*v1alpha1.RouteClaim{
				makeRouteClaim("some-host", "example.com", "/some-other-path"),
			},
			Routes: []*v1alpha1.Route{
				makeRoute("some-host", "example.com", "/some-path"),
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 2, len(v.Spec.HTTP))
				for i := range v.Spec.HTTP {
					testutil.AssertEqual(t, "HTTP Route len", 1, len(v.Spec.HTTP[i].Route))
					testutil.AssertEqual(t, "HTTP Route", networking.HTTPRouteDestination{
						Destination: networking.Destination{
							Host: resources.GatewayHost,
						},
						Weight: 100,
					}, v.Spec.HTTP[i].Route[0])
				}
			},
		},
		"Prefers Routes over Claims": {
			Claims: []*v1alpha1.RouteClaim{
				makeRouteClaim("some-host", "example.com", "/some-path"),
			},
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						AppName:         "some-app-name",
						RouteSpecFields: makeRouteSpecFields("some-host", "example.com", "/some-path"),
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 1, len(v.Spec.HTTP))
				testutil.AssertEqual(t, "HTTP Rewrite", &networking.HTTPRewrite{
					Authority: network.GetServiceHostname("some-app-name", ""),
				}, v.Spec.HTTP[0].Rewrite)
			},
		},
		"when there aren't any bound services, setup fault to 503": {
			Claims: []*v1alpha1.RouteClaim{
				makeRouteClaim("some-host", "example.com", "/some-path"),
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 1, len(v.Spec.HTTP))
				testutil.AssertEqual(t, "HTTP Fault", &networking.HTTPFaultInjection{
					Abort: &networking.InjectAbort{
						Percent:    100,
						HTTPStatus: http.StatusServiceUnavailable,
					},
				}, v.Spec.HTTP[0].Fault)
			},
		},
		"setup routes to bound services": {
			Claims: []*v1alpha1.RouteClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.RouteClaimSpec{
						RouteSpecFields: makeRouteSpecFields("some-host", "example.com", "/some-path"),
					},
				},
			},
			Routes: []*v1alpha1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: makeRouteSpecFields("some-host", "example.com", "/some-path"),
						AppName:         "ksvc-1",
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				expectedHTTP := []networking.HTTPRoute{
					networking.HTTPRoute{
						Match: []networking.HTTPMatchRequest{
							{URI: &istio.StringMatch{Regex: "^/some-path(/.*)?"}},
						},
						Rewrite: &networking.HTTPRewrite{
							Authority: network.GetServiceHostname("ksvc-1", "some-namespace"),
						},
						Route: []networking.HTTPRouteDestination{
							{
								Destination: networking.Destination{Host: "istio-ingressgateway.istio-system.svc.cluster.local"},
								Weight:      100,
							},
						},
						Headers: &networking.Headers{
							Request: &networking.HeaderOperations{
								Add: map[string]string{
									"X-Forwarded-Host": "some-host.example.com",
									"Forwarded":        "host=some-host.example.com",
								},
							},
						},
					},
				}

				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP", expectedHTTP, v.Spec.HTTP)
			},
		},
		"Hosts with subdomain": {
			Claims: []*v1alpha1.RouteClaim{
				makeRouteClaim("some-host", "example.com", "/some-path"),
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "Hosts", []string{"some-host.example.com"}, v.Spec.Hosts)
			},
		},
		"Hosts without subdomain": {
			Claims: []*v1alpha1.RouteClaim{
				makeRouteClaim("", "example.com", "/some-path"),
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "Hosts", []string{"example.com"}, v.Spec.Hosts)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			s, err := resources.MakeVirtualService(tc.Claims, tc.Routes)
			tc.Assert(t, s, err)
		})
	}
}

func ExampleMakeVirtualService() {
	claims := []*v1alpha1.RouteClaim{
		makeRouteClaim("some-host", "example.com/", ""),
	}

	routes := []*v1alpha1.Route{
		makeRoute("some-host", "example.com/", ""),
		makeRoute("some-host", "example.com/", "/some-path-1"),
		makeRoute("some-host", "example.com/", "/some-path-2"),
	}

	vs, err := resources.MakeVirtualService(claims, routes)
	if err != nil {
		panic(err)
	}

	for i, h := range vs.Spec.HTTP {
		fmt.Printf("Regex %d: %s\n", i, h.Match[0].URI.Regex)
	}

	// Output: Regex 0: ^(/.*)?
	// Regex 1: ^/some-path-1(/.*)?
	// Regex 2: ^/some-path-2(/.*)?
}
