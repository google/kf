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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fake *fake.FakeClient)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"listing routes fails": {
			ExpectedErr: errors.New("failed to fetch Routes: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().List("some-namespace")
			},
		},
		"display routes": {
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().List(gomock.Any()).Return([]v1alpha3.VirtualService{
					{
						Spec: v1alpha3.VirtualServiceSpec{
							Hosts: []string{"host-1", "host-2"},
							HTTP: []v1alpha3.HTTPRoute{
								{}, // nil Rewrite. Ensure we don't panic.
								{Rewrite: &v1alpha3.HTTPRewrite{Authority: "example.com"}}, // no subdomain.
								{Rewrite: &v1alpha3.HTTPRewrite{Authority: "app-1.example.com"}},
								{Rewrite: &v1alpha3.HTTPRewrite{Authority: "app-2.example.com"}, Match: []v1alpha3.HTTPMatchRequest{{URI: nil}, {URI: &v1alpha1.StringMatch{}}, {URI: &v1alpha1.StringMatch{Prefix: "/path1"}}}},
							},
						},
					},
				}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"host-1", "host-2", "app-1", "app-2", "path1"})
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
			cmd := routes.NewRoutesCommand(
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
				return
			}

			if tc.BufferF != nil {
				tc.BufferF(t, &buffer)
			}

			ctrl.Finish()
		})
	}
}
