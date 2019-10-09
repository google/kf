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

package apps

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		namespace string
		args      []string

		wantErr error
		setup   func(t *testing.T, fc *fake.FakeClient)
	}{
		"valid sync call": {
			namespace: "some-namespace",
			args:      []string{"some-app"},

			setup: func(t *testing.T, fc *fake.FakeClient) {
				fc.EXPECT().DeleteInForeground("some-namespace", "some-app")
				fc.EXPECT().WaitForDeletion(gomock.Any(), "some-namespace", "some-app", gomock.Any())
			},
		},
		"valid async call": {
			namespace: "some-namespace",
			args:      []string{"--async", "some-app"},

			setup: func(t *testing.T, fc *fake.FakeClient) {
				fc.EXPECT().DeleteInForeground("some-namespace", "some-app")
			},
		},
		"delete app error": {
			namespace: "some-namespace",
			args:      []string{"some-app"},

			setup: func(t *testing.T, fc *fake.FakeClient) {
				fc.EXPECT().DeleteInForeground(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			wantErr: errors.New("some error"),
		},
		"bad namespace error": {
			args: []string{"some-app"},

			setup: func(t *testing.T, fc *fake.FakeClient) {
				// expect no calls
			},
			wantErr: errors.New(utils.EmptyNamespaceError),
		},
		"wait error": {
			namespace: "some-namespace",
			args:      []string{"some-app"},

			setup: func(t *testing.T, fc *fake.FakeClient) {
				fc.EXPECT().DeleteInForeground(gomock.Any(), gomock.Any())
				fc.EXPECT().WaitForDeletion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
			},
			wantErr: errors.New("couldn't delete: some error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeDeleter := fake.NewFakeClient(ctrl)

			tc.setup(t, fakeDeleter)

			buffer := &bytes.Buffer{}
			c := NewDeleteCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeDeleter)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			gotErr := c.Execute()

			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}
		})
	}
}
