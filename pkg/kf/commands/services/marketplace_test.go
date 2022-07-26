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
	"strings"
	"testing"

	"bytes"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/v2/pkg/kf/commands/services"
	"github.com/google/kf/v2/pkg/kf/marketplace"
	"github.com/google/kf/v2/pkg/kf/marketplace/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewMarketplaceCommand(t *testing.T) {

	longDescription := strings.Repeat("X", 500)
	mockClusterBroker := &v1alpha1.ClusterServiceBroker{}
	mockClusterBroker.Name = "fake-broker"
	mockClusterBroker.Status = v1alpha1.CommonServiceBrokerStatus{
		Services: []v1alpha1.ServiceOffering{
			{
				DisplayName: "fake-service",
				Description: "fake-description",
				UID:         "00000000-0000-0000-0000-000000000000",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "fake-plan", Description: "description"},
					{DisplayName: "long-plan", Description: longDescription},
				},
			},
		},
	}

	mockMarketplace := &marketplace.KfMarketplace{}
	mockMarketplace.Brokers = append(mockMarketplace.Brokers, mockClusterBroker)

	cases := map[string]struct {
		Args  []string
		Setup func(t *testing.T, f *fake.FakeClientInterface)
		Space string

		ExpectedErr     error
		ExpectedStrings []string
	}{
		"too many params": {
			Args:        []string{"mydb"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"command params get passed correctly": {
			Args:  []string{},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), "custom-ns").Return(&marketplace.KfMarketplace{}, nil)
			},
		},
		"empty namespace": {
			Args:        []string{},
			ExpectedErr: errors.New(config.EmptySpaceError),
		},
		"command output outputs instance info": {
			Args:  []string{},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)
			},
			ExpectedStrings: []string{"fake-service", "fake-description", "fake-broker"},
		},
		"command output outputs plan info": {
			Args:  []string{"--service=fake-service"},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)
			},
			ExpectedStrings: []string{"fake-plan", "description"},
		},
		"command output outputs plan info for long description": {
			Args:  []string{"--service=fake-service"},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)
			},
			ExpectedStrings: []string{"long-plan", longDescription},
		},
		"blank marketplace": {
			Args:  []string{},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(&marketplace.KfMarketplace{}, nil)
			},
			ExpectedStrings: []string{},
		},
		"bad server call": {
			Args:  []string{},
			Space: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			client := fake.NewFakeClientInterface(ctrl)
			if tc.Setup != nil {
				tc.Setup(t, client)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
			}

			cmd := servicescmd.NewMarketplaceCommand(p, client)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
		})
	}
}
