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

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestCreateQuotaCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace   string
		quotaName   string
		wantErr     error
		wantErrStrs []string
		args        []string
		setup       func(t *testing.T, fakeCreator *fake.FakeClient)
		assert      func(t *testing.T, buffer *bytes.Buffer)
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"minimal config": {
			args: []string{"some-space"},
			setup: func(t *testing.T, fakeCreator *fake.FakeClient) {
				fakeCreator.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"successfully created", "quota", "some-space"})
			},
		},
		"some flags": {
			args: []string{"some-space", "-m", "1024M", "-c", "30"},
			setup: func(t *testing.T, fakeCreator *fake.FakeClient) {
				fakeCreator.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"successfully created", "quota", "some-space"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeCreator := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeCreator)
			}

			buffer := &bytes.Buffer{}

			c := NewCreateQuotaCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeCreator)
			c.SetOutput(buffer)

			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.wantErrStrs != nil {
				testutil.AssertErrorContainsAll(t, gotErr, tc.wantErrStrs)
				return
			}

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			testutil.AssertNil(t, "Command err", gotErr)
			ctrl.Finish()
		})
	}
}
