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

package routes_test

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/internal/routeutil"
	"github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
)

func TestCreateRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace string
		Args      []string
		Setup     func(t *testing.T, fake *fake.FakeClient)
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"some-space", "example.com", "extra", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts between 1 and 2 arg(s), received 3"), err)
			},
		},
		"namespace and space are different": {
			Namespace: "other-space",
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("SPACE (argument) and namespace (if provided) must match"), err)
			},
		},
		"creating route fails": {
			Args: []string{"some-space", "example.com", "--hostname=some-hostname"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create Route: some-error"), err)
			},
		},
		"missing hostname flag": {
			Args: []string{"some-space", "example.com"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("--hostname is required"), err)
			},
		},
		"namespace": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"uses namespace if SPACE is omitted": {
			Args:      []string{"example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"displays warning message if using space": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buffer.String(), []string{"deprecated"})
			},
		},
		"creates route with hostname and path": {
			Args: []string{"some-space", "example.com", "--hostname=some-hostname", "--path=somepath"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Create(gomock.Any(), &networking.VirtualService{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "networking.istio.io/v1alpha3",
						Kind:       "VirtualService",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: routeutil.EncodeRouteName("some-hostname", "example.com", "/somepath"),
						Annotations: map[string]string{
							"domain":   "example.com",
							"hostname": "some-hostname",
							"path":     "/somepath",
						},
					},
					Spec: networking.VirtualServiceSpec{
						Gateways: []string{"knative-ingress-gateway.knative-serving.svc.cluster.local"},
						Hosts:    []string{"some-hostname.example.com"},
						HTTP: []networking.HTTPRoute{
							{
								Match: []networking.HTTPMatchRequest{
									{
										URI: &v1alpha1.StringMatch{
											Prefix: "/somepath",
										},
									},
								},
								Route: []networking.HTTPRouteDestination{
									{
										Destination: networking.Destination{
											Host: "istio-ingressgateway.istio-system.svc.cluster.local",
											Port: networking.PortSelector{
												Number: 80,
											},
										},
										Weight: 100,
									},
								},
								Fault: &networking.HTTPFaultInjection{
									Abort: &networking.InjectAbort{
										Percent:    100,
										HTTPStatus: http.StatusServiceUnavailable,
									},
								},
							},
						},
					},
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewCreateRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				fake,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()

			if tc.Assert != nil {
				tc.Assert(t, &buffer, gotErr)
			}

			if gotErr != nil {
				return
			}
			ctrl.Finish()
		})
	}
}
