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
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestMapRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		Args        []string
		Setup       func(t *testing.T, appsfake *appsfake.FakeClient)
		ExpectedErr error
	}{
		"wrong number of args": {
			Args:        []string{"some-app", "example.com", "extra"},
			ExpectedErr: errors.New("accepts 2 arg(s), received 3"),
		},
		"transforming App fails": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
			ExpectedErr: errors.New("failed to map Route: some-error"),
		},
		"namespace": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().
					Transform("some-space", gomock.Any(), gomock.Any())
				appsfake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), "some-space", gomock.Any(), gomock.Any())
			},
		},
		"app name": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().
					Transform(gomock.Any(), "some-app", gomock.Any())
				appsfake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), "some-app", gomock.Any())
			},
		},
		"without namespace": {
			Args:        []string{"some-app", "example.com"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"transform App by adding new routes": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						oldApp := v1alpha1.App{}
						testutil.AssertNil(t, "err", m(&oldApp))

						testutil.AssertEqual(t, "Hostname", "some-host", oldApp.Spec.Routes[0].Hostname)
						testutil.AssertEqual(t, "Domain", "example.com", oldApp.Spec.Routes[0].Domain)
						testutil.AssertEqual(t, "Path", "/some-path", oldApp.Spec.Routes[0].Path)
					})
				appsfake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"transform App and keep old routes": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						oldApp := v1alpha1.App{}
						oldApp.Spec.Routes = []v1alpha1.RouteSpecFields{
							{Domain: "other.example.com"},
						}
						testutil.AssertNil(t, "err", m(&oldApp))

						// Existing Route
						testutil.AssertEqual(t, "Domain", "other.example.com", oldApp.Spec.Routes[0].Domain)
					})
				appsfake.EXPECT().WaitForConditionRoutesReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			appsfake := appsfake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, appsfake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewMapRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				appsfake,
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
