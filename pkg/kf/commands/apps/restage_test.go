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
	"github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestRestage(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"restages app": {
			Space: "default",
			Args:  []string{"my-app"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Restage(gomock.Any(), "default", "my-app")
				fake.EXPECT().DeployLogsForApp(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"no app name": {
			Space:       "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"restage app fails": {
			Space:       "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("failed to restage App: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Restage(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"restages app async": {
			Space: "default",
			Args:  []string{"--async", "my-app"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Restage(gomock.Any(), "default", "my-app")
			},
		},
		"restages app deployment fail": {
			Space:       "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("failed to restage App: some-log-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Restage(gomock.Any(), "default", "my-app")
				fake.EXPECT().DeployLogsForApp(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some-log-error"))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
			}

			cmd := NewRestageCommand(p, fake)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
			testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)

		})
	}
}
