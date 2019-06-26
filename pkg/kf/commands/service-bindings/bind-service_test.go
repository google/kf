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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewBindServiceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE", `--config={"ram_gb":4}`, "--binding-name=BINDING_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.CreateOption) {
					config := servicebindings.CreateOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{"ram_gb": 4.0}, config.Params())
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
		},
		"defaults config": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.CreateOption) {
					config := servicebindings.CreateOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{}, config.Params())
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
		},
		"bad config path": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE", `--config=/some/bad/path`},
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"writes binding info": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
			ExpectedStrings: []string{"APP_NAME", "SERVICE_INSTANCE"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewBindServiceCommand)
		})
	}
}
