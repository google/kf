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

package system

import (
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
)

type fakeServiceLister struct {
	services []*corev1.Service
	err      error
}

func (f *fakeServiceLister) List(selector labels.Selector) (ret []*corev1.Service, err error) {
	return f.services, f.err
}

func TestGetClusterIngresses(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		lister          fakeServiceLister
		expectIngresses []corev1.LoadBalancerIngress
		expectErr       error
	}{
		"server-error": {
			lister: fakeServiceLister{
				services: nil,
				err:      errors.New("some-server-error"),
			},
			expectErr: errors.New("some-server-error"),
		},
		"empty services": {
			lister: fakeServiceLister{
				services: []*corev1.Service{
					{},
				},
			},
		},
		"populated ingress": {
			lister: fakeServiceLister{
				services: []*corev1.Service{
					{
						Status: corev1.ServiceStatus{
							LoadBalancer: corev1.LoadBalancerStatus{
								Ingress: []corev1.LoadBalancerIngress{
									{IP: "8.8.8.8"},
									{Hostname: "kf.google.com"},
								},
							},
						},
					},
				},
			},
			expectIngresses: []corev1.LoadBalancerIngress{
				{IP: "8.8.8.8"},
				{Hostname: "kf.google.com"},
			},
		},
		"multiple matches": {
			lister: fakeServiceLister{
				services: []*corev1.Service{
					{
						Status: corev1.ServiceStatus{
							LoadBalancer: corev1.LoadBalancerStatus{
								Ingress: []corev1.LoadBalancerIngress{
									{IP: "8.8.8.8"},
									{Hostname: "kf.google.com"},
								},
							},
						},
					},
					{
						Status: corev1.ServiceStatus{
							LoadBalancer: corev1.LoadBalancerStatus{
								Ingress: []corev1.LoadBalancerIngress{
									{IP: "8.8.4.4"},
								},
							},
						},
					},
				},
			},
			expectIngresses: []corev1.LoadBalancerIngress{
				{IP: "8.8.8.8"},
				{Hostname: "kf.google.com"},
				{IP: "8.8.4.4"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ingresses, actualErr := GetClusterIngresses(&tc.lister)
			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "ingresses", tc.expectIngresses, ingresses)
		})
	}
}

func TestExtractProxyIngressFromList(t *testing.T) {
	cases := map[string]struct {
		ingresses []corev1.LoadBalancerIngress

		expectErr     error
		expectIngress string
	}{
		"zero ingresses": {
			expectErr: errors.New("no ingresses were found"),
		},
		"ingresses without IP": {
			ingresses: []corev1.LoadBalancerIngress{{}},
			expectErr: errors.New("no ingresses had IP addresses listed"),
		},
		"ingress with IP": {
			ingresses: []corev1.LoadBalancerIngress{
				{IP: "8.8.8.8"},
			},
			expectIngress: "8.8.8.8",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			in, actualErr := ExtractProxyIngressFromList(tc.ingresses)

			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "ingress", tc.expectIngress, in)
		})
	}
}
