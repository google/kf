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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestUnmapRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		Args        []string
		Setup       func(t *testing.T, fake *fake.FakeClient)
		ExpectedErr error
	}{
		"wrong number of args": {
			Args:        []string{"some-app", "example.com", "extra"},
			ExpectedErr: errors.New("accepts 2 arg(s), received 3"),
		},
		"transforming App fails": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
			ExpectedErr: errors.New("failed to unmap Route: some-error"),
		},
		"namespace": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform("some-space", gomock.Any(), gomock.Any())
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
		},
		"without namespace": {
			Args:        []string{"some-app", "example.com"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"App name": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), "some-app", gomock.Any())
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), "some-app", gomock.Any())
			},
		},
		"async does not wait": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path", "--async"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), "some-app", gomock.Any())
			},
		},
		"remove non-existent app": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_, _ string, m apps.Mutator) (*v1alpha1.App, error) {
						app := buildApp("some-app", "some-host", "example.com", "")
						return nil, m(&app)
					})
			},
			ExpectedErr: errors.New("failed to unmap Route: App some-app not found"),
		},
		"remove 1 of 1 routes": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						app := buildApp("some-app", "some-host", "example.com", "some-path")
						testutil.AssertNil(t, "err", m(&app))
						testutil.AssertEqual(t, "len", 0, len(app.Spec.Routes))
					})
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"remove 1 of 2 routes": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						app := v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteSpecFields{
							{Hostname: "some-host", Domain: "example.com", Path: "some-path"},
							{Hostname: "some-other-host", Domain: "example.com", Path: "some-path"},
						}

						testutil.AssertNil(t, "err", m(&app))
						testutil.AssertEqual(t, "routes", "some-other-host", app.Spec.Routes[0].Hostname)
					})
				fake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fake := fake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewUnmapRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
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
