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
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/envutil"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestUnsetEnvCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace       string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"wrong number of params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"unsetting variables fails": {
			Args:        []string{"app-name", "NAME"},
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), "app-name", gomock.Any()).Return(errors.New("some-error"))
			},
		},
		"custom namespace": {
			Args:      []string{"app-name", "NAME"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform("some-namespace", "app-name", gomock.Any())
			},
		},
		"unsets values": {
			Args: []string{"app-name", "NAME"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), "app-name", gomock.Any()).Do(func(ns, app string, mutator apps.Mutator) {
					input := apps.NewKfApp()
					input.SetEnvVars(envutil.MapToEnvVars(map[string]string{"NAME": "FOO", "PORT": "8080"}))
					svc := input.ToService()

					err := mutator(svc)
					testutil.AssertNil(t, "mutator err", err)

					result := (*apps.KfApp)(svc)
					actualVars := envutil.EnvVarsToMap(result.GetEnvVars())
					testutil.AssertEqual(t, "final values", map[string]string{"PORT": "8080"}, actualVars)
				})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.Namespace,
			}

			cmd := NewUnsetEnvCommand(p, fake)
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
