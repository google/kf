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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/cfutil"
	cfutilfake "github.com/google/kf/pkg/kf/cfutil/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	"github.com/google/kf/pkg/kf/commands/utils"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

func TestNewVcapServicesCommand(t *testing.T) {
	type serviceTest struct {
		Args      []string
		Setup     func(t *testing.T, f *fake.FakeClientInterface, systemEnvInjector *cfutilfake.FakeSystemEnvInjector)
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}

	runTest := func(t *testing.T, tc serviceTest, newCommand func(
		p *config.KfParams,
		client servicebindings.ClientInterface,
		systemEnvInjector cfutil.SystemEnvInjector,
	) *cobra.Command) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := fake.NewFakeClientInterface(ctrl)
		systemEnvInjector := cfutilfake.NewFakeSystemEnvInjector(ctrl)

		if tc.Setup != nil {
			tc.Setup(t, client, systemEnvInjector)
		}

		buf := new(bytes.Buffer)
		p := &config.KfParams{
			Namespace: tc.Namespace,
		}

		cmd := newCommand(p, client, systemEnvInjector)
		cmd.SetOutput(buf)
		cmd.SetArgs(tc.Args)
		_, actualErr := cmd.ExecuteC()
		if tc.ExpectedErr != nil || actualErr != nil {
			testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
			return
		}

		testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
	}
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface, systemEnvInjector *cfutilfake.FakeSystemEnvInjector) {
				f.EXPECT().List(gomock.Any(), gomock.Any()).Do(func(opts ...servicebindings.ListOption) {
					config := servicebindings.ListOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return([]apiv1beta1.ServiceBinding{}, nil)
				systemEnvInjector.EXPECT().GetVcapServices(gomock.Any(), gomock.Any()).Return([]cfutil.VcapService{}, nil)

			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"bad server call": {
			Args:      []string{"APP_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface, systemEnvInjector *cfutilfake.FakeSystemEnvInjector) {
				f.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewVcapServicesCommand)
		})
	}
}
