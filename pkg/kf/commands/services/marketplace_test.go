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

	"bytes"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/marketplace"
	"github.com/google/kf/pkg/kf/marketplace/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
)

func TestNewMarketplaceCommand(t *testing.T) {
	cases := map[string]struct {
		Args      []string
		Setup     func(t *testing.T, f *fake.FakeClientInterface)
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}{
		"too many params": {
			Args:        []string{"mydb"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"command params get passed correctly": {
			Args:      []string{},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace("custom-ns").Return(&marketplace.KfMarketplace{}, nil)
			},
		},
		"empty namespace": {
			Args:        []string{},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"command output outputs instance info": {
			Args:      []string{},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				fakeService := &v1beta1.ClusterServiceClass{}
				fakeService.Name = "00000000-0000-0000-0000-000000000000"
				fakeService.Spec.ExternalName = "fake-service"
				fakeService.Spec.Description = "fake-description"
				fakeService.Spec.ClusterServiceBrokerName = "fake-broker"

				f.EXPECT().Marketplace(gomock.Any()).Return(&marketplace.KfMarketplace{
					Services: []servicecatalog.Class{fakeService},
					Plans:    []servicecatalog.Plan{},
				}, nil)
			},
			ExpectedStrings: []string{"fake-service", "fake-description", "fake-broker"},
		},
		"command output outputs plan info": {
			Args:      []string{"--service=fake-service"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				fakeService := &v1beta1.ClusterServiceClass{}
				fakeService.Name = "00000000-0000-0000-0000-000000000000"
				fakeService.Spec.ExternalName = "fake-service"

				fakePlan := &v1beta1.ClusterServicePlan{}
				fakePlan.Name = "fake-plan"
				fakePlan.Spec.ExternalName = "fake-plan"
				fakePlan.Spec.Description = "description"
				fakePlan.Spec.ClusterServiceClassRef.Name = fakeService.Name
				fakePlan.Spec.CommonServicePlanSpec.ExternalName = fakePlan.Name

				f.EXPECT().Marketplace(gomock.Any()).Return(&marketplace.KfMarketplace{
					Services: []servicecatalog.Class{fakeService},
					Plans:    []servicecatalog.Plan{fakePlan},
				}, nil)
			},
			ExpectedStrings: []string{"fake-plan", "description"},
		},
		"blank marketplace": {
			Args:      []string{},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any()).Return(&marketplace.KfMarketplace{}, nil)
			},
			ExpectedStrings: []string{},
		},
		"bad server call": {
			Args:      []string{},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := fake.NewFakeClientInterface(ctrl)
			if tc.Setup != nil {
				tc.Setup(t, client)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.Namespace,
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
