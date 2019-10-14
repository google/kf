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

	utils "github.com/google/kf/pkg/kf/internal/utils/cli"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/apps/fake"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
)

func TestNewUnbindServiceCommand(t *testing.T) {
	cases := map[string]appsTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().UnbindService("custom-ns", "APP_NAME", "SERVICE_INSTANCE")
			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"bad server call": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClient) {
				f.EXPECT().UnbindService(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runAppsTest(t, tc, servicebindingscmd.NewUnbindServiceCommand)
		})
	}
}
