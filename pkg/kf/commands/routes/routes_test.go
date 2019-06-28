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
	fakeapp "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	fakeroute "github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"listing routes fails": {
			ExpectedErr: errors.New("failed to fetch Routes: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient) {
				fakeRoute.EXPECT().List("some-namespace")
			},
		},
		"display routes": {
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					{
						Spec: v1alpha1.RouteSpec{
							Hostname: "host-1",
							Domain:   "example.com",
							Path:     "/path1",
						},
					},
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-1", "example.com", "/path1"})
			},
		},
		"display apps": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{
								"service-1",
								"service-2",
							},
						},
					},
				}, nil)

				fakeApp.EXPECT().Get("some-namespace", "service-1").Return(&serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "app-1",
					},
				}, nil)
				fakeApp.EXPECT().Get("some-namespace", "service-2").Return(&serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "app-2",
					},
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"app-1, app-2"})
			},
		},
		"fetching Knative Service fails": {
			Namespace:   "some-namespace",
			ExpectedErr: errors.New("fetching Knative Service failed: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroute.FakeClient, fakeApp *fakeapp.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{
								"service-1",
								"service-2",
							},
						},
					},
				}, nil)

				fakeApp.EXPECT().Get("some-namespace", "service-1").Return(nil, errors.New("some-error"))
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRoute := fakeroute.NewFakeClient(ctrl)
			fakeApp := fakeapp.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeRoute, fakeApp)
			}

			var buffer bytes.Buffer
			cmd := routes.NewRoutesCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				fakeRoute,
				fakeApp,
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
