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
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	fakeapps "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	fakerouteclaims "github.com/google/kf/pkg/kf/routeclaims/fake"
	fakeroutes "github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"listing routes fails": {
			Namespace:   "some-namespace",
			ExpectedErr: errors.New("failed to fetch Routes: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeApp.EXPECT().List(gomock.Any()).AnyTimes()
				fakeRouteClaim.EXPECT().List(gomock.Any()).AnyTimes()
				fakeRoute.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"listing route claims fails": {
			Namespace:   "some-namespace",
			ExpectedErr: errors.New("failed to fetch RouteClaims: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeRoute.EXPECT().List(gomock.Any()).AnyTimes()
				fakeApp.EXPECT().List(gomock.Any()).AnyTimes()
				fakeRouteClaim.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"listing apps fails": {
			Namespace:   "some-namespace",
			ExpectedErr: errors.New("failed to fetch Apps: some-error"),
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeRouteClaim.EXPECT().List(gomock.Any()).AnyTimes()
				fakeRoute.EXPECT().List(gomock.Any()).AnyTimes()
				fakeApp.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeRoute.EXPECT().List("some-namespace")
				fakeRouteClaim.EXPECT().List("some-namespace")
				fakeApp.EXPECT().List("some-namespace")
			},
		},
		"display routes": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeRouteClaim.EXPECT().List(gomock.Any())
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					buildRoute("host-1", "example.com", "/path1"),
					buildRoute("host-2", "example.com", "/path1"),
					buildRoute("host-3", "example.com", "/path2"),
				}, nil)
				fakeApp.EXPECT().List(gomock.Any()).Return([]v1alpha1.App{
					buildApp("app-1", "host-1", "example.com", "path1"),
					buildApp("app-2", "host-1", "example.com", "path1"),

					// Host doesn't match and should be in a different group
					buildApp("app-3", "host-2", "example.com", "/path1"),

					// Don't show deleted timestamp
					buildDeletedApp("deleted-app", "host-3", "example.com", "/path2"),
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-1", "example.com", "/path1", "app-1, app-2"})
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-2", "example.com", "/path1", "app-3"})
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-3", "example.com", "/path2"})

				// Ensure it doesn't show deleted-app
				if strings.Index(buffer.String(), "deleted-app") >= 0 {
					t.Fatal("should not have 'deleted-app'")
				}
			},
		},
		"display claim": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRoute *fakeroutes.FakeClient, fakeRouteClaim *fakerouteclaims.FakeClient, fakeApp *fakeapps.FakeClient) {
				fakeRouteClaim.EXPECT().List(gomock.Any()).Return([]v1alpha1.RouteClaim{
					buildRouteClaim("host-1", "example.com", "/path1"),
				}, nil)
				fakeRoute.EXPECT().List(gomock.Any()).Return([]v1alpha1.Route{
					buildRoute("host-2", "example.com", "/path2"),
				}, nil)
				fakeApp.EXPECT().List(gomock.Any()).Return([]v1alpha1.App{
					buildApp("app-2", "host-2", "example.com", "path2"),
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-1", "example.com", "/path1"})
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-2", "example.com", "/path2", "app-2"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRoute := fakeroutes.NewFakeClient(ctrl)
			fakeRouteClaim := fakerouteclaims.NewFakeClient(ctrl)
			fakeApp := fakeapps.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeRoute, fakeRouteClaim, fakeApp)
			}

			var buffer bytes.Buffer
			cmd := routes.NewRoutesCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				fakeRoute,
				fakeRouteClaim,
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

func buildApp(name, hostname, domain, path string) v1alpha1.App {
	return v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.AppSpec{
			Routes: []v1alpha1.RouteSpecFields{
				{
					Hostname: hostname,
					Domain:   domain,
					Path:     path,
				},
			},
		},
	}
}

func buildDeletedApp(name, hostname, domain, path string) v1alpha1.App {
	app := buildApp(name, hostname, domain, path)
	app.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	return app
}

func buildRoute(hostname, domain, path string) v1alpha1.Route {
	return v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			RouteSpecFields: v1alpha1.RouteSpecFields{
				Hostname: hostname,
				Domain:   domain,
				Path:     path,
			},
		},
	}
}

func buildRouteClaim(hostname, domain, path string) v1alpha1.RouteClaim {
	return v1alpha1.RouteClaim{
		Spec: v1alpha1.RouteClaimSpec{
			RouteSpecFields: v1alpha1.RouteSpecFields{
				Hostname: hostname,
				Domain:   domain,
				Path:     path,
			},
		},
	}
}
