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

package servicebindings_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

type appsCommandFactory func(p *config.KfParams, client apps.Client) *cobra.Command

type appsTest struct {
	Args      []string
	Setup     func(t *testing.T, f *fake.FakeClient)
	Namespace string

	ExpectedErr     error
	ExpectedStrings []string
}

func runAppsTest(t *testing.T, tc appsTest, newCommand appsCommandFactory) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := fake.NewFakeClient(ctrl)
	if tc.Setup != nil {
		tc.Setup(t, client)
	}

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Namespace: tc.Namespace,
	}

	cmd := newCommand(p, client)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
		return
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}

func TestNewBindServiceCommand(t *testing.T) {
	cases := map[string]appsTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE", `--config={"ram_gb":4}`, "--binding-name=BINDING_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().BindService("custom-ns", "APP_NAME", &v1alpha1.AppSpecServiceBinding{
					Instance:    "SERVICE_INSTANCE",
					Parameters:  json.RawMessage(`{"ram_gb":4}`),
					BindingName: "BINDING_NAME",
				})

				f.EXPECT().WaitForConditionServiceBindingsReadyTrue(gomock.Any(), "custom-ns", "APP_NAME", gomock.Any())
			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE", `--config={"ram_gb":4}`, "--binding-name=BINDING_NAME"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"defaults config": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().BindService("custom-ns", "APP_NAME", &v1alpha1.AppSpecServiceBinding{
					Instance:   "SERVICE_INSTANCE",
					Parameters: json.RawMessage(`{}`),
				})

				f.EXPECT().WaitForConditionServiceBindingsReadyTrue(gomock.Any(), "custom-ns", "APP_NAME", gomock.Any())
			},
		},
		"bad config path": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE", `--config=/some/bad/path`},
			Namespace:   "custom-ns",
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().BindService(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"async": {
			Args:      []string{"--async", "APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "default",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().BindService(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"failed binding": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().BindService(gomock.Any(), gomock.Any(), gomock.Any())
				f.EXPECT().WaitForConditionServiceBindingsReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("binding already exists"))
			},
			ExpectedErr: errors.New("bind failed: binding already exists"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runAppsTest(t, tc, servicebindingscmd.NewBindServiceCommand)
		})
	}
}
