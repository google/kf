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

package quotas

import (
	"bytes"
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/quotas/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/golang/mock/gomock"
)

func TestUpdateQuotaCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace   string
		quotaName   string
		wantErr     error
		args        []string
		setup       func(t *testing.T, fakeUpdater *fake.FakeClient)
		assert      func(t *testing.T, buffer *bytes.Buffer)
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"update error": {
			args:    []string{"some-quota", "-m", "100z"},
			wantErr: errors.New("some-error"),
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
		},
		"configured namespace": {
			args:      []string{"some-quota"},
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform("some-namespace", gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		"some flags": {
			args: []string{"some-quota", "-m", "1024M"},
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeUpdater := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeUpdater)
			}

			buffer := &bytes.Buffer{}

			c := NewUpdateQuotaCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeUpdater)
			c.SetOutput(buffer)

			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			testutil.AssertNil(t, "Command err", gotErr)
			testutil.AssertEqual(t, "SilenceUsage", true, c.SilenceUsage)

			ctrl.Finish()
		})
	}
}
