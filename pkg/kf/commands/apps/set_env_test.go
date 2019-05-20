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

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/golang/mock/gomock"
)

func TestSetEnvCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace       string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeEnvironmentClient)
	}{
		"wrong number of params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"setting variables fails": {
			Args:        []string{"app-name", "NAME", "VALUE"},
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Set("app-name", gomock.Any(), gomock.Any()).Return(errors.New("some-error"))
			},
		},
		"custom namespace": {
			Args:      []string{"app-name", "NAME", "VALUE"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(appName string, values map[string]string, opts ...kf.SetEnvOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", kf.SetEnvOptions(opts).Namespace())
				})
			},
		},
		"sets values": {
			Args: []string{"app-name", "NAME", "VALUE"},
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Set("app-name", map[string]string{"NAME": "VALUE"}, gomock.Any())
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeEnvironmentClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Output:    buf,
				Namespace: tc.Namespace,
			}

			cmd := NewSetEnvCommand(p, fake)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
			testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)

			ctrl.Finish()
		})
	}
}
