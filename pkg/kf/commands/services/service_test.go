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
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/services/fake"
)

func TestNewGetServiceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Get("custom-ns", "mydb").Return(dummyServerInstance("mydb-instance1"), nil)
			},
		},
		"empty namespace": {
			Args:        []string{"mydb"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"command output outputs instance info": {
			Args:      []string{"mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Get(gomock.Any(), "mydb").Return(dummyServerInstance("mydb-instance1"), nil)
			},
			ExpectedStrings: []string{"mydb-instance1"},
		},
		"service not found": {
			Args:      []string{"some-missing-service"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Get(gomock.Any(), "some-missing-service").Return(nil, nil)
			},
			ExpectedStrings: []string{"<empty>"},
		},
		"bad server call": {
			Args:      []string{"mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().Get(gomock.Any(), "mydb").Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewGetServiceCommand)
		})
	}
}
