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
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/commands/routes"
	"github.com/google/kf/v2/pkg/kf/testutil"

	corev1 "k8s.io/api/core/v1"
)

func TestNewProxyRouteCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		IngressGateways []corev1.LoadBalancerIngress
		ExpectedStrings []string
		ExpectedErr     error
	}{
		"no route": {
			Space:       "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"minimal configuration": {
			Space:           "default",
			Args:            []string{"myhost.example.com", "--no-start=true"},
			ExpectedStrings: []string{"myhost.example.com", "8.8.8.8"},
			IngressGateways: []corev1.LoadBalancerIngress{{IP: "8.8.8.8"}},
			ExpectedErr:     nil,
		},
		"autodetect failure": {
			Space:           "default",
			Args:            []string{"myhost.example.com", "--no-start=true"},
			IngressGateways: []corev1.LoadBalancerIngress{}, // no gateways
			ExpectedErr:     errors.New("no ingresses were found"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gomock.NewController(t)

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
				TargetSpace: &v1alpha1.Space{
					Status: v1alpha1.SpaceStatus{
						IngressGateways: tc.IngressGateways,
					},
				},
			}

			cmd := routes.NewProxyRouteCommand(p)
			cmd.SetOutput(buf)
			cmd.SetContext(configlogging.SetupLogger(context.Background(), buf))
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
