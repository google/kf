// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/ptr"
)

func TestStableDeduplicateSpaceDomainList(t *testing.T) {
	cases := map[string]struct {
		domains  []SpaceDomain
		expected []SpaceDomain
	}{
		"no entries": {
			domains:  nil,
			expected: nil,
		},
		"order is preserved": {
			domains: []SpaceDomain{
				{Domain: "bbb.com"},
				{Domain: "zzz.com"},
				{Domain: "aaa.com"},
			},
			expected: []SpaceDomain{
				{Domain: "bbb.com"},
				{Domain: "zzz.com"},
				{Domain: "aaa.com"},
			},
		},
		"order preserved based on first occurrence": {
			domains: []SpaceDomain{
				{Domain: "bbb.com"},
				{Domain: "zzz.com"},
				{Domain: "aaa.com"},
				{Domain: "aaa.com"},
				{Domain: "zzz.com"},
				{Domain: "bbb.com"},
			},
			expected: []SpaceDomain{
				{Domain: "bbb.com"},
				{Domain: "zzz.com"},
				{Domain: "aaa.com"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := StableDeduplicateSpaceDomainList(tc.domains)
			testutil.AssertEqual(t, "deduplicated domains", tc.expected, actual)
		})
	}
}

func TestSpaceStatus_FindIngressIP(t *testing.T) {
	cases := map[string]struct {
		gateways []corev1.LoadBalancerIngress
		expected *string
	}{
		"empty gateways": {
			gateways: nil,
			expected: nil,
		},
		"skips host load balancers": {
			gateways: []corev1.LoadBalancerIngress{
				{IP: "", Hostname: "my-lb-host.cloud.google.com"},
				{IP: "127.0.0.1"},
			},
			expected: ptr.String("127.0.0.1"),
		},
		"finds lexicographical first IP": {
			gateways: []corev1.LoadBalancerIngress{
				{IP: "8.8.8.8"},
				{IP: "8.8.4.4"},
			},
			expected: ptr.String("8.8.4.4"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := SpaceStatus{IngressGateways: tc.gateways}

			actual := status.FindIngressIP()
			testutil.AssertEqual(t, "IP", tc.expected, actual)
		})
	}
}

func ExampleSpace_DefaultDomainOrBlank() {
	s := Space{}
	fmt.Printf("No Domain: %q\n", s.DefaultDomainOrBlank())

	s.Status.NetworkConfig.Domains = []SpaceDomain{
		{Domain: "first.is.default.domain"},
		{Domain: "example.com"},
	}

	fmt.Printf("Domain: %q\n", s.DefaultDomainOrBlank())

	// Output: No Domain: ""
	// Domain: "first.is.default.domain"
}
