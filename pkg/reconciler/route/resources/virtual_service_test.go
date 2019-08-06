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
	"sort"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
	"github.com/knative/serving/pkg/network"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/kmeta"
)

func TestMakeVirtualService(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Routes []*v1alpha1.Route
		Assert func(t *testing.T, v *networking.VirtualService, err error)
	}{
		"empty list of routes": {
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertErrorsEqual(t, errors.New("routes must not be empty"), err)
			},
		},
		"proper Meta": {
			Routes: []*v1alpha1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some-namespace",
						Labels:    map[string]string{"a": "1", "b": "2"},
					},
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "some-path",
						},
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "TypeMeta", metav1.TypeMeta{
					APIVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
				}, v.TypeMeta)

				route := &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
					},
				}

				ownerRef := *kmeta.NewControllerRef(route)
				ownerRef.Controller = nil
				ownerRef.BlockOwnerDeletion = nil

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
					OwnerReferences: []metav1.OwnerReference{
						ownerRef,
					},
				}, v.ObjectMeta)
			},
		},
		"Path Matchers": {
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
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
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 1, len(v.Spec.HTTP))
				testutil.AssertEqual(t, "HTTP Route len", 1, len(v.Spec.HTTP[0].Route))
				testutil.AssertEqual(t, "HTTP Route", networking.HTTPRouteDestination{
					Destination: networking.Destination{
						Host: resources.GatewayHost,
					},
					Weight: 100,
				}, v.Spec.HTTP[0].Route[0])
			},
		},
		"when there aren't any bound services, setup fault to 503": {
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
					},
				},
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
			Routes: []*v1alpha1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
						AppName: "ksvc-1",
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "HTTP len", 1, len(v.Spec.HTTP))
				testutil.AssertEqual(t, "HTTP Rewrite", &networking.HTTPRewrite{
					Authority: network.GetServiceHostname("ksvc-1", "some-namespace"),
				}, v.Spec.HTTP[0].Rewrite)
				testutil.AssertEqual(t, "HTTP Match len", 1, len(v.Spec.HTTP[0].Match))
				testutil.AssertEqual(t, "HTTP Match", "^/some-path(/.*)?", v.Spec.HTTP[0].Match[0].URI.Regex)
			},
		},
		"Hosts with subdomain": {
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "some-host",
							Domain:   "example.com",
							Path:     "/some-path",
						},
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "Hosts len", 1, len(v.Spec.Hosts))
				testutil.AssertEqual(t, "Hosts", []string{"some-host.example.com"}, v.Spec.Hosts)
			},
		},
		"Hosts without subdomain": {
			Routes: []*v1alpha1.Route{
				{
					Spec: v1alpha1.RouteSpec{
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "",
							Domain:   "example.com",
							Path:     "/some-path",
						},
					},
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "Hosts len", 1, len(v.Spec.Hosts))
				testutil.AssertEqual(t, "Hosts", []string{"example.com"}, v.Spec.Hosts)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			s, err := resources.MakeVirtualService(tc.Routes)
			tc.Assert(t, s, err)
		})
	}
}

func ExampleMakeVirtualService() {
	vs1, err := resources.MakeVirtualService([]*v1alpha1.Route{
		{
			Spec: v1alpha1.RouteSpec{
				RouteSpecFields: v1alpha1.RouteSpecFields{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path-1",
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	vs2, err := resources.MakeVirtualService([]*v1alpha1.Route{
		{
			Spec: v1alpha1.RouteSpec{
				RouteSpecFields: v1alpha1.RouteSpecFields{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path-2",
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	r := algorithms.Merge(
		v1alpha1.HTTPRoutes(vs1.Spec.HTTP),
		v1alpha1.HTTPRoutes(vs2.Spec.HTTP),
	).(v1alpha1.HTTPRoutes)

	// Sort for display purposes
	sort.Sort(r)

	for i, h := range r {
		fmt.Printf("Regex %d: %s\n", i, h.Match[0].URI.Regex)
	}

	// Output: Regex 0: ^/some-path-1(/.*)?
	// Regex 1: ^/some-path-2(/.*)?
}
