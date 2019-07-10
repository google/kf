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
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"sort"
	"strings"
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

func TestVirtualServiceName_Deterministic(t *testing.T) {
	t.Parallel()

	r1 := resources.VirtualServiceName("host-1", "example1.com")
	r2 := resources.VirtualServiceName("host-1", "example1.com")
	r3 := resources.VirtualServiceName("host-2", "example1.com")
	r4 := resources.VirtualServiceName("host-1", "example2.com")

	testutil.AssertEqual(t, "r1 and r2", r1, r2)
	testutil.AssertEqual(t, "r1 and r2", r1, r2)

	for _, r := range []string{r3, r4} {
		if r1 == r {
			t.Fatalf("expected %s to not equal %s", r, r1)
		}
	}
}

func TestVirtualServiceName_ValidDNS(t *testing.T) {
	t.Parallel()

	// We'll use an instantiation of rand so we can seed it with 0 for
	// repeatable tests.
	rand := rand.New(rand.NewSource(0))
	randStr := func() string {
		buf := make([]byte, rand.Intn(128)+1)
		for i := range buf {
			buf[i] = byte(rand.Intn('z'-'a') + 'a')
		}
		return strings.ToUpper(path.Join("./", string(buf)))
	}

	history := map[string]bool{}

	validDNS := func(r string) {
		testutil.AssertRegexp(t, "valid DNS", `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`, r)
		testutil.AssertEqual(t, fmt.Sprintf("len: %d", len(r)), true, len(r) <= 64)
		testutil.AssertEqual(t, "collison", false, history[r])
	}

	for i := 0; i < 10000; i++ {
		r := resources.VirtualServiceName(randStr(), randStr())
		validDNS(r)
		history[r] = true
	}

	// Empty name
	validDNS(resources.VirtualServiceName())

	// Only non-alphanumeric characters
	validDNS(resources.VirtualServiceName(".", "-", "$"))
}

func TestMakeVirtualService(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Route  *v1alpha1.Route
		Assert func(t *testing.T, v *networking.VirtualService, err error)
	}{
		"proper Meta": {
			Route: &v1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Labels:    map[string]string{"a": "1", "b": "2"},
				},
				Spec: v1alpha1.RouteSpec{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "some-path",
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
						Hostname: "some-host",
						Domain:   "example.com",
						Path:     "/some-path",
					},
				}

				ownerRef := *kmeta.NewControllerRef(route)
				ownerRef.Controller = nil
				ownerRef.BlockOwnerDeletion = nil

				testutil.AssertEqual(t, "ObjectMeta", metav1.ObjectMeta{
					Name:      resources.VirtualServiceName(route.Spec.Hostname, route.Spec.Domain),
					Namespace: "some-namespace",
					Labels:    map[string]string{"a": "1", "b": "2"},
					Annotations: map[string]string{
						"domain":   "example.com",
						"hostname": "some-host",
						"path":     "/some-path",
					},
					OwnerReferences: []metav1.OwnerReference{
						ownerRef,
					},
				}, v.ObjectMeta)
			},
		},
		"Path Matchers": {
			Route: &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path",
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
			Route: &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path",
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
			Route: &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path",
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
			Route: &v1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.RouteSpec{
					Hostname:            "some-host",
					Domain:              "example.com",
					Path:                "/some-path",
					KnativeServiceNames: []string{"ksvc-1"},
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
			Route: &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					Hostname: "some-host",
					Domain:   "example.com",
					Path:     "/some-path",
				},
			},
			Assert: func(t *testing.T, v *networking.VirtualService, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "Hosts len", 1, len(v.Spec.Hosts))
				testutil.AssertEqual(t, "Hosts", []string{"some-host.example.com"}, v.Spec.Hosts)
			},
		},
		"Hosts without subdomain": {
			Route: &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					Hostname: "",
					Domain:   "example.com",
					Path:     "/some-path",
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
			s, err := resources.MakeVirtualService(tc.Route)
			tc.Assert(t, s, err)
		})
	}
}

func ExampleMakeVirtualService() {
	vs1, err := resources.MakeVirtualService(&v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			Hostname: "some-host",
			Domain:   "example.com",
			Path:     "/some-path-1",
		},
	})
	if err != nil {
		panic(err)
	}

	vs2, err := resources.MakeVirtualService(&v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			Hostname: "some-host",
			Domain:   "example.com",
			Path:     "/some-path-2",
		},
	})
	if err != nil {
		panic(err)
	}

	r := algorithms.Merge(
		resources.HTTPRoutes(vs1.Spec.HTTP),
		resources.HTTPRoutes(vs2.Spec.HTTP),
	).(resources.HTTPRoutes)

	// Sort for display purposes
	sort.Sort(r)

	for i, h := range r {
		fmt.Printf("Regex %d: %s\n", i, h.Match[0].URI.Regex)
	}

	// Output: Regex 0: ^/some-path-1(/.*)?
	// Regex 1: ^/some-path-2(/.*)?
}
