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

package kf_test

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type fakeDependencies struct {
	apiserver *testutil.FakeApiServer
}

func TestIstioClient_ListIngresses(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		opts  []kf.ListIngressesOption
		setup func(mockK8s kubernetes.Interface)

		expectErr error
	}{
		"server-error": {
			opts: []kf.ListIngressesOption{
				kf.WithListIngressesService("bad-service"),
			},
			expectErr: errors.New(`services "bad-service" not found`),
		},
		"default values": {
			setup: func(mockK8s kubernetes.Interface) {
				svc := &corev1.Service{}
				svc.Name = "istio-ingressgateway"
				mockK8s.CoreV1().Services("istio-system").Create(svc)
			},
		},
		"custom values": {
			opts: []kf.ListIngressesOption{
				kf.WithListIngressesNamespace("custom-ns"),
				kf.WithListIngressesService("custom-gateway"),
			},
			setup: func(mockK8s kubernetes.Interface) {
				svc := &corev1.Service{}
				svc.Name = "custom-gateway"
				mockK8s.CoreV1().Services("custom-ns").Create(svc)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			mockK8s := testclient.NewSimpleClientset()
			if tc.setup != nil {
				tc.setup(mockK8s)
			}
			client := kf.NewIstioClient(mockK8s)

			ingresses, actualErr := client.ListIngresses(tc.opts...)
			if actualErr != nil || tc.expectErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
				return
			}

			testutil.AssertNotNil(t, "ingresses", ingresses)
		})
	}
}

func TestExtractIngressFromList(t *testing.T) {
	cases := map[string]struct {
		ingresses []corev1.LoadBalancerIngress
		err       error

		expectErr     error
		expectIngress string
	}{
		"error identity": {
			err:       errors.New("test-err"),
			expectErr: errors.New("test-err"),
		},
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
			in, actualErr := kf.ExtractIngressFromList(tc.ingresses, tc.err)

			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "ingress", tc.expectIngress, in)
		})
	}
}
