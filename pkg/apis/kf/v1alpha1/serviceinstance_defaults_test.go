// Copyright 2020 Google LLC
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

package v1alpha1

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestServiceInstance_SetDefaults(t *testing.T) {
	cases := testutil.ApisDefaultableTestSuite{
		"empty": {
			Context: defaultContext(),
			Input:   &ServiceInstance{},
			Want:    &ServiceInstance{},
		},
		"custom-config": {
			Context: sampleConfig(),
			Input:   &ServiceInstance{},
			Want:    &ServiceInstance{},
		},
	}

	cases.Run(t)
}

func TestServiceType_SetDefaults(t *testing.T) {
	cases := testutil.ApisDefaultableTestSuite{
		"UPS empty": {
			Context: defaultContext(),
			Input: &ServiceType{
				UPS: &UPSInstance{},
			},
			Want: &ServiceType{
				UPS: &UPSInstance{},
			},
		},
		"UPS custom-config": {
			Context: sampleConfig(),
			Input: &ServiceType{
				UPS: &UPSInstance{},
			},
			Want: &ServiceType{
				UPS: &UPSInstance{},
			},
		},
		"brokered empty": {
			Context: defaultContext(),
			Input: &ServiceType{
				Brokered: &BrokeredInstance{},
			},
			Want: &ServiceType{
				Brokered: &BrokeredInstance{},
			},
		},
		"brokered custom-config": {
			Context: sampleConfig(),
			Input: &ServiceType{
				Brokered: &BrokeredInstance{},
			},
			Want: &ServiceType{
				Brokered: &BrokeredInstance{},
			},
		},
		"osb empty": {
			Context: sampleConfig(),
			Input: &ServiceType{
				OSB: &OSBInstance{},
			},
			Want: &ServiceType{
				OSB: &OSBInstance{
					ProgressDeadlineSeconds: DefaultServiceInstanceProgressDeadlineSeconds,
				},
			},
		},
		"osb custom timeout": {
			Context: sampleConfig(),
			Input: &ServiceType{
				OSB: &OSBInstance{
					ProgressDeadlineSeconds: 30,
				},
			},
			Want: &ServiceType{
				OSB: &OSBInstance{
					ProgressDeadlineSeconds: 30,
				},
			},
		},
	}

	cases.Run(t)
}
