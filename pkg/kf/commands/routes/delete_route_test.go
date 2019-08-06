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
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	fakeapp "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/commands/utils"
	fakerouteclaims "github.com/google/kf/pkg/kf/routeclaims/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestDeleteRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace string
		Args      []string
		Setup     func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient)
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"example.com", "extra"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts 1 arg(s), received 2"), err)
			},
		},
		"listing apps fails": {
			Args:      []string{"example.com"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to list apps: some-error"), err)
			},
		},
		"unmaping route fails": {
			Args:      []string{"example.com"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]v1alpha1.App{{}}, nil)
				fakeApps.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to unmap Route: some-error"), err)
			},
		},
		"deleting route fails": {
			Args:      []string{"example.com"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().List(gomock.Any(), gomock.Any())
				fakeRouteClaims.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to delete Route: some-error"), err)
			},
		},
		"namespace": {
			Args:      []string{"example.com"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					List("some-namespace", gomock.Any()).
					Return([]v1alpha1.App{{}}, nil)
				fakeApps.EXPECT().
					Transform("some-namespace", gomock.Any(), gomock.Any())
				fakeRouteClaims.EXPECT().
					Delete("some-namespace", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"without namespace": {
			Args: []string{"example.com"},
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					List("some-namespace", gomock.Any())
				fakeApps.EXPECT().
					Transform("some-namespace", gomock.Any(), gomock.Any())
				fakeRouteClaims.EXPECT().
					Delete("some-namespace", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(utils.EmptyNamespaceError), err)
			},
		},
		"delete RouteClaim": {
			Args:      []string{"example.com", "--hostname=some-hostname", "--path=somepath"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fakeRouteClaims *fakerouteclaims.FakeClient, fakeApps *appsfake.FakeClient) {
				fakeApps.EXPECT().
					List(gomock.Any(), gomock.Any())
				expectedName := v1alpha1.GenerateRouteClaimName(
					"some-hostname",
					"example.com",
					"/somepath",
				)
				fakeRouteClaims.EXPECT().
					Delete(
						gomock.Any(),
						expectedName,
					)
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRouteClaims := fakerouteclaims.NewFakeClient(ctrl)
			fakeApps := fakeapp.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeRouteClaims, fakeApps)
			}

			var buffer bytes.Buffer
			cmd := routes.NewDeleteRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				fakeRouteClaims,
				fakeApps,
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
