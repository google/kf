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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/routes"
	routesfake "github.com/google/kf/v2/pkg/kf/routes/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Space  string
		Args   []string
		Setup  func(t *testing.T, routesfake *routesfake.FakeClient)
		Assert func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"some-space", "example.com", "extra", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts between 1 and 2 arg(s), received 3"), err)
			},
		},
		"space and space are different": {
			Space: "other-space",
			Args:  []string{"some-space", "example.com", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(`SPACE (argument="some-space") and space (flag="other-space") (if provided) must match`), err)
			},
		},
		"missing hostname flag": {
			Args:  []string{"some-space", "example.com"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				routesfake.EXPECT().Create(gomock.Any(), "some-space", gomock.Any())
				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"creating route fails": {
			Args:  []string{"some-space", "example.com", "--hostname=some-hostname"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				routesfake.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create Route: some-error"), err)
			},
		},
		"space": {
			Args:  []string{"some-space", "example.com", "--hostname=some-hostname"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				routesfake.EXPECT().Create(gomock.Any(), "some-space", gomock.Any())
				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"without space": {
			Args: []string{"some-space", "example.com", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(config.EmptySpaceError), err)
			},
		},
		"space is default and space is not": {
			Space: "default",
			Args:  []string{"some-space", "example.com", "--hostname=some-hostname"},
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				routesfake.EXPECT().Create(gomock.Any(), "some-space", gomock.Any())
				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"uses space if SPACE is omitted": {
			Args:  []string{"example.com", "--hostname=some-hostname"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				routesfake.EXPECT().Create(gomock.Any(), "some-space", gomock.Any())
				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"creates route with hostname and path": {
			Args:  []string{"some-space", "example.com", "--hostname=some-hostname", "--path=somepath"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				expectName := v1alpha1.GenerateRouteName(
					"some-hostname",
					"example.com",
					"/somepath",
				)

				routesfake.EXPECT().Create(gomock.Any(), gomock.Any(),
					&v1alpha1.Route{
						TypeMeta: metav1.TypeMeta{
							Kind: "Route",
						},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "some-space",
							Name:      expectName,
						},
						Spec: v1alpha1.RouteSpec{
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
								Path:     "/somepath",
							},
						},
					},
				)

				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", expectName, gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"creates route path but with missing hostname": {
			Args:  []string{"some-space", "example.com", "--path=somepath"},
			Space: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient) {
				expectName := v1alpha1.GenerateRouteName(
					"",
					"example.com",
					"/somepath",
				)

				routesfake.EXPECT().Create(gomock.Any(), gomock.Any(),
					&v1alpha1.Route{
						TypeMeta: metav1.TypeMeta{
							Kind: "Route",
						},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "some-space",
							Name:      expectName,
						},
						Spec: v1alpha1.RouteSpec{
							RouteSpecFields: v1alpha1.RouteSpecFields{
								Hostname: "",
								Domain:   "example.com",
								Path:     "/somepath",
							},
						},
					},
				)

				routesfake.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "some-space", expectName, gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			routesfake := routesfake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, routesfake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewCreateRouteCommand(
				&config.KfParams{
					Space: tc.Space,
				},
				routesfake,
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

		})
	}
}
