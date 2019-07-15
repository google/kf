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

package builds

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/sources/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"knative.dev/pkg/apis"
)

func TestNewListBuildsCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		args      []string
		namespace string
		setup     func(t *testing.T, fakeSources *fake.FakeClient)

		wantErr         error
		expectedStrings []string
	}{
		"invalid number of args": {
			args:    []string{"asdf"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"missing namespace": {
			args:    []string{},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"no contents": {
			namespace: "my-ns",
			setup: func(t *testing.T, fakeSources *fake.FakeClient) {
				list := []v1alpha1.Source{}
				fakeSources.
					EXPECT().
					List("my-ns").
					Return(list, nil)
			},
			expectedStrings: []string{"Name", "Age", "Ready", "Reason"},
		},
		"contents": {
			namespace: "my-ns",
			setup: func(t *testing.T, fakeSources *fake.FakeClient) {
				bld := v1alpha1.Source{}
				bld.Name = "my-build"
				bld.Status.Conditions = []apis.Condition{{
					Type:   "Succeeded",
					Status: "TESTING",
					Reason: "SomeMessage",
				}}
				bld.Status.Image = "gcr.io/my-image"

				list := []v1alpha1.Source{bld}
				fakeSources.
					EXPECT().
					List("my-ns").
					Return(list, nil)
			},
			expectedStrings: []string{"my-build", "TESTING", "SomeMessage", "gcr.io/my-image"},
		},
		"server failure": {
			namespace: "my-ns",
			setup: func(t *testing.T, fakeSources *fake.FakeClient) {
				fakeSources.
					EXPECT().
					List("my-ns").
					Return(nil, errors.New("some-server-error"))
			},
			wantErr: errors.New("some-server-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSources := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeSources)
			}

			buffer := &bytes.Buffer{}

			c := NewListBuildsCommand(&config.KfParams{Namespace: tc.namespace}, fakeSources)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertContainsAll(t, buffer.String(), tc.expectedStrings)

			ctrl.Finish()
		})
	}
}
