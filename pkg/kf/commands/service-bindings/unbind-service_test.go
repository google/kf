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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/apps/fake"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewUnbindServiceCommand(t *testing.T) {
	type serviceTest struct {
		Args      []string
		Setup     func(t *testing.T, f *fake.FakeClient)
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}

	runTest := func(t *testing.T, tc serviceTest, newCommand func(p *config.KfParams, client apps.Client) *cobra.Command) {
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
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&v1alpha1.App{}, nil)
				f.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(namespace string, app *v1alpha1.App) (*v1alpha1.App, error) {

					config := servicebindings.DeleteOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())

					return &v1alpha1.App{}, nil
				})
			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"defaults config": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Delete("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.DeleteOption) {
					config := servicebindings.DeleteOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(nil)
			},
		},
		"bad server call": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewUnbindServiceCommand)
		})
	}
}
