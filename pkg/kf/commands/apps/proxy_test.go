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

package apps

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fakeapps "github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"

	corev1 "k8s.io/api/core/v1"
)

func TestNewProxyCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		IngressGateways []corev1.LoadBalancerIngress
		Setup           func(t *testing.T, lister *fakeapps.FakeClient)
	}{
		"no app name": {
			Space:       "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"no url for app": {
			Space:           "default",
			Args:            []string{"my-app", "--no-start=true"},
			ExpectedErr:     errors.New("no public routes for App my-app"),
			IngressGateways: []corev1.LoadBalancerIngress{{IP: "8.8.8.8"}},
			Setup: func(t *testing.T, lister *fakeapps.FakeClient) {
				lister.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Status: v1alpha1.AppStatus{
						Routes: nil,
					},
				}, nil)
			},
		},
		"only wildcard for app": {
			Space:           "default",
			Args:            []string{"my-app", "--no-start=true"},
			ExpectedErr:     errors.New("couldn't find suitable App domain"),
			IngressGateways: []corev1.LoadBalancerIngress{{IP: "8.8.8.8"}},
			Setup: func(t *testing.T, lister *fakeapps.FakeClient) {
				lister.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Status: v1alpha1.AppStatus{
						Routes: []v1alpha1.AppRouteStatus{
							{
								QualifiedRouteBinding: v1alpha1.QualifiedRouteBinding{
									Source: v1alpha1.RouteSpecFields{
										Hostname: "*",
										Domain:   "example.com",
									},
								},
							},
						},
					},
				}, nil)
			},
		},
		"minimal configuration": {
			Space:           "default",
			Args:            []string{"my-app", "--no-start=true"},
			ExpectedErr:     nil,
			IngressGateways: []corev1.LoadBalancerIngress{{IP: "8.8.8.8"}},
			Setup: func(t *testing.T, lister *fakeapps.FakeClient) {
				lister.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Status: v1alpha1.AppStatus{
						Routes: []v1alpha1.AppRouteStatus{
							{
								QualifiedRouteBinding: v1alpha1.QualifiedRouteBinding{
									Source: v1alpha1.RouteSpecFields{
										Hostname: "my-app",
										Domain:   "example.com",
									},
								},
							},
						},
					},
				}, nil)
			},
		},
		"autodetect failure": {
			Space:           "default",
			Args:            []string{"my-app"},
			ExpectedErr:     errors.New("no ingresses were found"),
			IngressGateways: []corev1.LoadBalancerIngress{}, // no gateways
			Setup: func(t *testing.T, lister *fakeapps.FakeClient) {
				lister.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Status: v1alpha1.AppStatus{
						Routes: []v1alpha1.AppRouteStatus{
							{
								QualifiedRouteBinding: v1alpha1.QualifiedRouteBinding{
									Source: v1alpha1.RouteSpecFields{
										Hostname: "my-app",
										Domain:   "example.com",
										Path:     "/",
									},
								},
							},
						},
					},
				}, nil)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeAppClient := fakeapps.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeAppClient)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
				TargetSpace: &v1alpha1.Space{
					Status: v1alpha1.SpaceStatus{
						IngressGateways: tc.IngressGateways,
					},
				},
			}

			cmd := NewProxyCommand(p, fakeAppClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
			testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)

		})
	}
}
