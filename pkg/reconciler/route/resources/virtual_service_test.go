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
	"math/rand"
	"net/http"
	"regexp"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/kmeta"
)

func TestEncodeRouteName_Deterministic(t *testing.T) {
	t.Parallel()

	r1 := resources.VirtualServiceName("host-1", "example1.com", "somePath1")
	r2 := resources.VirtualServiceName("host-1", "example1.com", "somePath1")
	r3 := resources.VirtualServiceName("host-2", "example1.com", "somePath1")
	r4 := resources.VirtualServiceName("host-1", "example2.com", "somePath1")
	r5 := resources.VirtualServiceName("host-1", "example1.com", "somePath2")

	testutil.AssertEqual(t, "r1 and r2", r1, r2)
	testutil.AssertEqual(t, "r1 and r2", r1, r2)

	for _, r := range []string{r3, r4, r5} {
		if r1 == r {
			t.Fatalf("expected %s to not equal %s", r, r1)
		}
	}
}

func TestEncodeRouteName_ValidDNS(t *testing.T) {
	t.Parallel()

	// We'll use an instantiation of rand so we can seed it with 0 for
	// repeatable tests.
	rand := rand.New(rand.NewSource(0))
	randStr := func() string {
		buf := make([]byte, rand.Intn(19)+1)
		for i := range buf {
			buf[i] = byte(rand.Intn('z'-'a') + 'a')
		}
		return string(buf)
	}

	pattern := regexp.MustCompile(`[^a-z0-9-_]`)
	history := map[string]bool{}

	// Basically we're going to try a mess of different things to ensure that
	// certain rules are followed:
	// [a-z0-9_-]
	for i := 0; i < 10000; i++ {
		r := resources.VirtualServiceName(randStr(), randStr(), randStr())
		testutil.AssertEqual(t, "invalid rune: "+r, false, pattern.MatchString(r))

		testutil.AssertEqual(t, "collison", false, history[r])
		history[r] = true
	}
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

				testutil.AssertEqual(t, "ObjectMeta", metav1.ObjectMeta{
					Name:      resources.VirtualServiceName(route.Spec.Hostname, route.Spec.Domain, route.Spec.Path),
					Namespace: "some-namespace",
					Labels:    map[string]string{"a": "1", "b": "2"},
					Annotations: map[string]string{
						"domain":   "example.com",
						"hostname": "some-host",
						"path":     "/some-path",
					},
					OwnerReferences: []metav1.OwnerReference{
						*kmeta.NewControllerRef(route),
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
						Prefix: "/some-path",
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
						Port: networking.PortSelector{
							Number: 80,
						},
					},
					Weight: 100,
				}, v.Spec.HTTP[0].Route[0])
			},
		},
		"Setup fault to 503": {
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
