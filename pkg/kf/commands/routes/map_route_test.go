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
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/commands/utils"
	clientroutes "github.com/google/kf/pkg/kf/routes"
	routesfake "github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestMapRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace string
		Args      []string
		Setup     func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient)
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"some-app", "example.com", "extra"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts 2 arg(s), received 3"), err)
			},
		},
		"fetching app fails": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to fetch app: some-error"), err)
			},
		},
		"transforming Route fails": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&serving.Service{}, nil)
				routesfake.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to map Route: some-error"), err)
			},
		},
		"namespace": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get("some-space", gomock.Any()).Return(&serving.Service{}, nil)
				routesfake.EXPECT().Upsert("some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"without namespace": {
			Args: []string{"some-app", "example.com"},
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get("some-space", gomock.Any()).Return(&serving.Service{}, nil)
				routesfake.EXPECT().Upsert("some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(utils.EmptyNamespaceError), err)
			},
		},
		"fetches app": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get(gomock.Any(), "some-app").Return(&serving.Service{}, nil)
				routesfake.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"transform Route": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-app",
						UID:  types.UID("some-uid"),
					},
				}, nil)
				routesfake.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ string, newR *v1alpha1.Route, m clientroutes.Merger) {
					testutil.AssertEqual(t, "name",
						resources.VirtualServiceName(
							"some-host",
							"example.com",
							"/some-path",
						),
						newR.Name,
					)
					testutil.AssertEqual(t, "Spec.Hostname", "some-host", newR.Spec.Hostname)
					testutil.AssertEqual(t, "Spec.Domain", "example.com", newR.Spec.Domain)
					testutil.AssertEqual(t, "Spec.Path", "/some-path", newR.Spec.Path)

					oldR := v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{"some-other-app"},
						},
					}
					m(newR, &oldR)
					testutil.AssertEqual(t, "names", []string{"some-app", "some-other-app"}, newR.Spec.KnativeServiceNames)
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"don't re-add app": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, appsfake *appsfake.FakeClient) {
				appsfake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&serving.Service{}, nil)
				routesfake.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ string, newR *v1alpha1.Route, m clientroutes.Merger) {
					oldR := v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{"some-app"},
						},
					}
					m(&oldR, newR)
					testutil.AssertEqual(t, "names ", []string{"some-app"}, newR.Spec.KnativeServiceNames)
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			routesfake := routesfake.NewFakeClient(ctrl)
			appsfake := appsfake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, routesfake, appsfake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewMapRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				routesfake,
				appsfake,
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
