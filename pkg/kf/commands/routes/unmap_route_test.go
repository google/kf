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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/commands/utils"
	clientroutes "github.com/google/kf/pkg/kf/routes"
	"github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
)

func TestUnmapRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace string
		Args      []string
		Setup     func(t *testing.T, fake *fake.FakeClient)
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"some-app", "example.com", "extra"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts 2 arg(s), received 3"), err)
			},
		},
		"transforming Route fails": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to unmap Route: some-error"), err)
			},
		},
		"namespace": {
			Args:      []string{"some-app", "example.com"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform("some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"without namespace": {
			Args: []string{"some-app", "example.com"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform("some-space", gomock.Any(), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(utils.EmptyNamespaceError), err)
			},
		},
		"Route name": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), resources.VirtualServiceName("some-host", "example.com", "/some-path"), gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"remove non-existent app": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _ string, m clientroutes.Mutator) {
					r := v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{"some-other-app"},
						},
					}
					testutil.AssertEqual(t, "err", errors.New("App some-app not found"), m(&r))
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"remove 1 of 1 apps": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _ string, m clientroutes.Mutator) {
					r := v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{"some-app"},
						},
					}
					testutil.AssertNil(t, "err", m(&r))
					testutil.AssertEqual(t, "names len", 0, len(r.Spec.KnativeServiceNames))
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"remove 1 of 2 apps": {
			Args:      []string{"some-app", "example.com", "--hostname=some-host", "--path=some-path"},
			Namespace: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _ string, m clientroutes.Mutator) {
					r := v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							KnativeServiceNames: []string{"some-other-app", "some-app"},
						},
					}
					testutil.AssertNil(t, "err", m(&r))
					testutil.AssertEqual(t, "names ", []string{"some-other-app"}, r.Spec.KnativeServiceNames)
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
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
					Namespace: tc.Namespace,
				},
				fake,
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
