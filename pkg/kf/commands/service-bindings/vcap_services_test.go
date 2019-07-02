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

func TestNewVcapServicesCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().GetVcapServices("APP_NAME", gomock.Any()).Do(func(app string, opts ...servicebindings.GetVcapServicesOption) {
					config := servicebindings.GetVcapServicesOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(nil, nil)
			},
		},
		"bad server call": {
			Args: []string{"APP_NAME"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().GetVcapServices(gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
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
