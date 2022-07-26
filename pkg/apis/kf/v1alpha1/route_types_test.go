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
	"context"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/ptr"
)

func ExampleRouteSpecFields_String() {
	r := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
		Path:     "bar",
	}

	fmt.Println(r.String())

	// Output: foo.example.com/bar
}

func ExampleRouteSpecFields_String_without_hostname() {
	r := RouteSpecFields{
		Domain: "example.com",
		Path:   "bar",
	}

	fmt.Println(r.String())

	// Output: example.com/bar
}

func ExampleRouteSpecFields_String_without_path() {
	r := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
	}

	fmt.Println(r.String())

	// Output: foo.example.com
}

func ExampleRouteSpecFields_IsWildcard() {
	example := RouteSpecFields{Hostname: "example"}
	fmt.Println("Example is wildcard:", example.IsWildcard())

	star := RouteSpecFields{Hostname: "*", Domain: "example.com"}
	fmt.Println("Star is wildcard:", star.IsWildcard())

	// Output: Example is wildcard: false
	// Star is wildcard: true
}

func ExampleRouteSpecFields_Host() {
	r := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
	}

	fmt.Println(r.Host())

	// Output: foo.example.com
}

func ExampleRouteSpecFields_Host_noHostname() {
	r := RouteSpecFields{
		Domain: "example.com",
	}

	fmt.Println(r.Host())

	// Output: example.com
}

func ExampleRouteSpecFields_ToURL() {
	url := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
		Path:     "bar",
	}.ToURL()

	fmt.Println((&url).String())

	// Output: //foo.example.com/bar
}

func TestRouteWeightBinding_Merge(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		original   RouteWeightBinding
		toMerge    RouteWeightBinding
		wantWeight int32
	}{
		"empty": {
			original:   RouteWeightBinding{},
			toMerge:    RouteWeightBinding{},
			wantWeight: 2,
		},
		"original filled": {
			original: RouteWeightBinding{
				Weight: ptr.Int32(3),
			},
			toMerge:    RouteWeightBinding{},
			wantWeight: 4,
		},
		"merge filled": {
			original: RouteWeightBinding{},
			toMerge: RouteWeightBinding{
				Weight: ptr.Int32(3),
			},
			wantWeight: 4,
		},
		"both filled": {
			original: RouteWeightBinding{
				Weight: ptr.Int32(5),
			},
			toMerge: RouteWeightBinding{
				Weight: ptr.Int32(3),
			},
			wantWeight: 8,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.original.Merge(tc.toMerge)

			testutil.AssertEqual(t, "weight", tc.wantWeight, *tc.original.Weight)
		})
	}
}

// EqualsBinding tests equality between two bindings.
func ExampleRouteWeightBinding_EqualsBinding() {
	first := RouteWeightBinding{}
	first.Hostname = "*"
	first.Domain = "example.com"
	first.Path = "/some/path"
	first.DestinationPort = ptr.Int32(8080)

	second := *first.DeepCopy()
	fmt.Println("same port equal?", first.EqualsBinding(context.Background(), second))

	second.DestinationPort = ptr.Int32(9999)
	fmt.Println("different ports equal?", first.EqualsBinding(context.Background(), second))

	// Output: same port equal? true
	// different ports equal? false
}

func TestMergeBindings(t *testing.T) {
	t.Parallel()

	rwb := func(domain string, weight *int32) RouteWeightBinding {
		return RouteWeightBinding{
			RouteSpecFields: RouteSpecFields{
				Hostname: "host",
				Domain:   domain,
				Path:     "/",
			},
			Weight: weight,
		}
	}

	cases := map[string]struct {
		in   []RouteWeightBinding
		want []RouteWeightBinding
	}{
		"nil is identity": {
			in:   nil,
			want: nil,
		},
		"noop": {
			in: []RouteWeightBinding{
				rwb("test", nil),
				rwb("other", ptr.Int32(4)),
			},
			want: []RouteWeightBinding{
				rwb("test", nil),
				rwb("other", ptr.Int32(4)),
			},
		},
		"multiple merges": {
			in: []RouteWeightBinding{
				rwb("test", nil),
				rwb("test", ptr.Int32(2)),
				rwb("test", ptr.Int32(4)),
			},
			want: []RouteWeightBinding{
				rwb("test", ptr.Int32(7)),
			},
		},
		"merge preserves first encounter order": {
			in: []RouteWeightBinding{
				rwb("a", nil),
				rwb("b", nil),
				rwb("c", nil),
				rwb("c", nil),
				rwb("b", nil),
				rwb("a", nil),
			},
			want: []RouteWeightBinding{
				rwb("a", ptr.Int32(2)),
				rwb("b", ptr.Int32(2)),
				rwb("c", ptr.Int32(2)),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := MergeBindings(tc.in)
			testutil.AssertEqual(t, "merged", tc.want, got)
		})
	}
}

func TestRoute_IsOrphaned(t *testing.T) {
	cases := map[string]struct {
		generation  int64
		observedGen int64
		bindings    []RouteDestination

		wantOrphaned bool
	}{
		"blank is not orphaned": {
			generation:   0,
			observedGen:  0,
			bindings:     nil,
			wantOrphaned: false,
		},
		"matched generation with nil bindings is orphaned": {
			generation:   42,
			observedGen:  42,
			bindings:     nil,
			wantOrphaned: true,
		},
		"matched generation with empty bindings is orphaned": {
			generation:   42,
			observedGen:  42,
			bindings:     []RouteDestination{},
			wantOrphaned: true,
		},
		"mismatch generation is not orphaned": {
			generation:   42,
			observedGen:  0,
			bindings:     nil,
			wantOrphaned: false,
		},
		"matched generation with routes is not orphaned": {
			generation:   1,
			observedGen:  1,
			bindings:     []RouteDestination{{}},
			wantOrphaned: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			route := Route{}
			route.Generation = tc.generation
			route.Status.ObservedGeneration = tc.observedGen
			route.Status.Bindings = tc.bindings

			gotOrphaned := route.IsOrphaned()
			testutil.AssertEqual(t, "IsOrphaned", tc.wantOrphaned, gotOrphaned)
		})
	}
}

func TestRoute_hasDestination(t *testing.T) {
	exists := RouteDestination{
		ServiceName: "exists",
		Port:        80,
		Weight:      1,
	}

	doesNotExist := RouteDestination{
		ServiceName: "doesNotExist",
		Port:        80,
		Weight:      1,
	}

	existsRoute := Route{}
	existsRoute.Status.PropagateBindings([]RouteDestination{exists})

	cases := map[string]struct {
		route       Route
		destination RouteDestination
		wantExists  bool
	}{
		"blank route": {
			route:       Route{},
			destination: doesNotExist,
			wantExists:  false,
		},
		"populated route missing dest": {
			route:       existsRoute,
			destination: doesNotExist,
			wantExists:  false,
		},
		"populated route with dest": {
			route:       existsRoute,
			destination: exists,
			wantExists:  true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotExists := tc.route.hasDestination(tc.destination)
			testutil.AssertEqual(t, "hasDestination", tc.wantExists, gotExists)
		})
	}
}

func TestRouteWeightBinding_Qualify(t *testing.T) {
	const defaultDomain = "sample.domain"
	const serviceName = "my-app"

	cases := map[string]struct {
		binding RouteWeightBinding

		wantQualified QualifiedRouteBinding
	}{
		"empty binding": {
			binding: RouteWeightBinding{},
			wantQualified: QualifiedRouteBinding{
				Source: RouteSpecFields{
					Domain: defaultDomain,
				},
				Destination: RouteDestination{
					Port:        DefaultRouteDestinationPort,
					ServiceName: serviceName,
					Weight:      defaultRouteWeight,
				},
			},
		},
		"full binding": {
			binding: RouteWeightBinding{
				DestinationPort: ptr.Int32(9999),
				Weight:          ptr.Int32(33),
				RouteSpecFields: RouteSpecFields{
					Hostname: "host",
					Domain:   "some.domain",
					Path:     "/some/path",
				},
			},
			wantQualified: QualifiedRouteBinding{
				Source: RouteSpecFields{
					Hostname: "host",
					Domain:   "some.domain",
					Path:     "/some/path",
				},
				Destination: RouteDestination{
					Port:        9999,
					ServiceName: serviceName,
					Weight:      33,
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotQualified := tc.binding.Qualify(defaultDomain, serviceName)

			testutil.AssertEqual(t, "qualified", tc.wantQualified, gotQualified)
		})
	}
}

func ExampleQualifiedRouteBinding_ToUnqualified() {
	qrb := QualifiedRouteBinding{
		Source: RouteSpecFields{
			Hostname: "host",
			Domain:   "some.domain",
			Path:     "/some/path",
		},
		Destination: RouteDestination{
			Port:        9999,
			ServiceName: "my-service",
			Weight:      33,
		},
	}

	unqualified := qrb.ToUnqualified()

	fmt.Println("URL", unqualified.String())
	fmt.Println("Port", *unqualified.DestinationPort)
	fmt.Println("Weight", *unqualified.Weight)

	// Output: URL host.some.domain/some/path
	// Port 9999
	// Weight 33
}

func TestMergeQualifiedBindings(t *testing.T) {
	t.Parallel()

	qrb := func(domain string, weight int32) QualifiedRouteBinding {
		return QualifiedRouteBinding{
			Source: RouteSpecFields{
				Hostname: "host",
				Domain:   domain,
				Path:     "/",
			},
			Destination: RouteDestination{
				ServiceName: "app",
				Port:        int32(DefaultRouteDestinationPort),
				Weight:      weight,
			},
		}
	}

	cases := map[string]struct {
		in   []QualifiedRouteBinding
		want []QualifiedRouteBinding
	}{
		"nil is identity": {
			in:   nil,
			want: nil,
		},
		"noop": {
			in: []QualifiedRouteBinding{
				qrb("test", 1),
				qrb("other", 4),
			},
			want: []QualifiedRouteBinding{
				qrb("test", 1),
				qrb("other", 4),
			},
		},
		"multiple merges": {
			in: []QualifiedRouteBinding{
				qrb("test", 1),
				qrb("test", 2),
				qrb("test", 4),
			},
			want: []QualifiedRouteBinding{
				qrb("test", 7),
			},
		},
		"merge preserves first encounter order": {
			in: []QualifiedRouteBinding{
				qrb("a", 1),
				qrb("b", 1),
				qrb("c", 1),
				qrb("c", 1),
				qrb("b", 1),
				qrb("a", 1),
			},
			want: []QualifiedRouteBinding{
				qrb("a", 2),
				qrb("b", 2),
				qrb("c", 2),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := MergeQualifiedBindings(tc.in)
			testutil.AssertEqual(t, "merged", tc.want, got)
		})
	}
}
