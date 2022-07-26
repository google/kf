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
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/commands/routes"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

func TestUnmapRoute(t *testing.T) {
	t.Parallel()

	someSpace := &v1alpha1.Space{}

	for tn, tc := range map[string]struct {
		Space       string
		TargetSpace *v1alpha1.Space

		Args        []string
		Setup       func(t *testing.T, fake *fake.FakeClient)
		ExpectedErr error
	}{
		"wrong number of args": {
			Args:        []string{"some-app", "example.com", "extra"},
			ExpectedErr: errors.New("accepts 2 arg(s), received 3"),
		},
		"transforming App fails": {
			Args:        []string{"some-app", "example.com"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
			ExpectedErr: errors.New("failed to unmap Route: some-error"),
		},
		"space": {
			Args:        []string{"some-app", "example.com"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
		},
		"without space": {
			Args:        []string{"some-app", "example.com"},
			ExpectedErr: errors.New(config.EmptySpaceError),
		},
		"App name": {
			Args:        []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), "some-app", gomock.Any())
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), "some-app", gomock.Any())
			},
		},
		"async does not wait": {
			Args:        []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path", "--async"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), "some-app", gomock.Any())
			},
		},
		"remove 1 of 1 routes": {
			Args:        []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, _, _ string, m apps.Mutator) {
						app := buildApp("some-app", "some-host", "example.com", "some-path")
						testutil.AssertNil(t, "err", m(&app))
						testutil.AssertEqual(t, "len", 0, len(app.Spec.Routes))
					})
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"remove 1 of 2 routes": {
			Args:        []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Space:       "some-space",
			TargetSpace: someSpace,
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, _, _ string, m apps.Mutator) {
						app := v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteWeightBinding{
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Hostname: "some-host",
									Domain:   "example.com",
									Path:     "some-path",
								},
							},
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Hostname: "some-other-host",
									Domain:   "example.com",
									Path:     "some-path",
								},
							},
						}
						for _, r := range app.Spec.Routes {
							app.Status.Routes = append(
								app.Status.Routes,
								v1alpha1.AppRouteStatus{
									QualifiedRouteBinding: r.Qualify("default.domain", app.Name),
								},
							)
						}

						testutil.AssertNil(t, "err", m(&app))
						testutil.AssertEqual(t, "len", 1, len(app.Spec.Routes))
						testutil.AssertEqual(t, "routes", "some-other-host", app.Spec.Routes[0].Hostname)
					})
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
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
			cmd := routes.NewUnmapRouteCommand(
				&config.KfParams{
					Space:       tc.Space,
					TargetSpace: tc.TargetSpace,
				},
				fake,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
			}
		})
	}
}

func buildApp(name, hostname, domain, path string) v1alpha1.App {
	route := v1alpha1.RouteWeightBinding{
		Weight: ptr.Int32(1),
		RouteSpecFields: v1alpha1.RouteSpecFields{
			Hostname: hostname,
			Domain:   domain,
			Path:     path,
		},
	}

	return v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.AppSpec{
			Routes: []v1alpha1.RouteWeightBinding{
				route,
			},
		},
		Status: v1alpha1.AppStatus{
			Routes: []v1alpha1.AppRouteStatus{
				{
					QualifiedRouteBinding: route.Qualify("default.domain", name),
				},
			},
		},
	}
}
