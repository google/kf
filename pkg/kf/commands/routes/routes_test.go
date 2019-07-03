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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	fakeroute "github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fakeRoute *fakeroute.FakeClient)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"listing routes fails": {
			Namespace:   "some-namespace",
			ExpectedErr: errors.New("failed to fetch Routes: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient) {
				fakeRoute.EXPECT().List("some-namespace")
			},
		},
		"display routes": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					{
						Spec: v1alpha1.RouteSpec{
							Hostname:            "host-1",
							Domain:              "example.com",
							Path:                "/path1",
							KnativeServiceNames: []string{"app-1", "app-2"},
						},
					},
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-1", "example.com", "/path1", "app-1, app-2"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRoute := fakeroute.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeRoute)
			}

			var buffer bytes.Buffer
			cmd := routes.NewRoutesCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				fakeRoute,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			if tc.BufferF != nil {
				tc.BufferF(t, &buffer)
			}

			ctrl.Finish()
		})
	}
}
