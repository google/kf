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

package services_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/services/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewCreateServiceCommand(t *testing.T) {

	cases := map[string]serviceTest{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService("mydb", "db-service", "free", gomock.Any()).Do(func(instance, service, plan string, opts ...services.CreateServiceOption) {
					config := services.CreateServiceOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{"ram_gb": 4.0}, config.Params())
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(dummyServerInstance("mydb"), nil)
			},
		},
		"empty namespace": {
			Args:        []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"defaults config": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService("mydb", "db-service", "free", gomock.Any()).Do(func(instance, service, plan string, opts ...services.CreateServiceOption) {
					config := services.CreateServiceOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{}, config.Params())
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(dummyServerInstance("mydb"), nil)
			},
		},
		"bad path": {
			Args:        []string{"db-service", "free", "mydb", `--config=/some/bad/path`},
			Namespace:   "custom-ns",
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewCreateServiceCommand)
		})
	}
}
