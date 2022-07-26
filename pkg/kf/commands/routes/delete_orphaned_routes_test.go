// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routes

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	fakeroutes "github.com/google/kf/v2/pkg/kf/routes/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestDeleteOrphanedRoutes(t *testing.T) {
	t.Parallel()

	buildRoute := func(ns, name string, orphaned bool) v1alpha1.Route {
		rc := v1alpha1.Route{}
		rc.Namespace = ns
		rc.Name = name
		rc.Generation = 10
		rc.Status.ObservedGeneration = rc.Generation
		if !orphaned {
			rc.Status.Bindings = []v1alpha1.RouteDestination{{}}
		}

		return rc
	}

	cases := map[string]struct {
		space   string
		args    []string
		setup   func(t *testing.T, fakeRoutes *fakeroutes.FakeClient)
		wantErr error
	}{
		"wrong number of args": {
			args:    []string{"example.com"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"listing routes fails": {
			space: "some-namespace",
			setup: func(t *testing.T, fakeRoutes *fakeroutes.FakeClient) {
				fakeRoutes.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			wantErr: errors.New("failed to list Routes: some-error"),
		},
		"deleting route fails": {
			space: "some-namespace",
			setup: func(t *testing.T, fakeRoutes *fakeroutes.FakeClient) {
				fakeRoutes.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.Route{
					buildRoute("some-namespace", "should-fail", true),
				}, nil)

				fakeRoutes.
					EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("deletion failed"))
			},
			wantErr: errors.New("failed to delete Route: deletion failed"),
		},
		"deletes routes": {
			space: "some-namespace",
			setup: func(t *testing.T, fakeRoutes *fakeroutes.FakeClient) {
				fakeRoutes.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.Route{
					buildRoute("some-namespace", "orphaned", true),
					buildRoute("some-namespace", "not-orphaned", false),
					buildRoute("some-namespace", "orphaned2", true),
				}, nil)
				fakeRoutes.EXPECT().Delete(gomock.Any(), "some-namespace", "orphaned")
				fakeRoutes.EXPECT().Delete(gomock.Any(), "some-namespace", "orphaned2")
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRoutes := fakeroutes.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeRoutes)
			}

			var buffer bytes.Buffer
			cmd := NewDeleteOrphanedRoutesCommand(
				&config.KfParams{
					Space: tc.space,
				},
				fakeRoutes,
			)

			if tc.args == nil {
				// We have to set this to something that is non-nil so that
				// cobra doesn't go use os.Args (which might have test flags
				// set).
				tc.args = make([]string, 0)
			}

			cmd.SetArgs(tc.args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

		})
	}
}
