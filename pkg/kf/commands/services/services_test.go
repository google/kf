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

	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/testutil"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/services/fake"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func TestNewServicesCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"too many params": {
			Args:        []string{"foo", "bar"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 2"),
		},
		"custom namespace": {
			Namespace: "test-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().ListServices(gomock.Any()).
					DoAndReturn(func(opts ...services.ListServicesOption) (*v1beta1.ServiceInstanceList, error) {
						options := services.ListServicesOptions(opts)
						testutil.AssertEqual(t, "namespace", "test-ns", options.Namespace())

						return &v1beta1.ServiceInstanceList{}, nil
					})
			},
		},
		"empty result": {
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				emptyList := &v1beta1.ServiceInstanceList{Items: []v1beta1.ServiceInstance{}}
				f.EXPECT().ListServices(gomock.Any()).Return(emptyList, nil)
			},
			ExpectedErr: nil, // explicitly expecting no failure with zero length list
		},
		"full result": {
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				serviceList := &v1beta1.ServiceInstanceList{Items: []v1beta1.ServiceInstance{
					*dummyServerInstance("service-1"),
					*dummyServerInstance("service-2"),
				}}
				f.EXPECT().ListServices(gomock.Any()).Return(serviceList, nil)
			},
			ExpectedStrings: []string{"service-1", "service-2"},
		},
		"bad server call": {
			ExpectedErr: errors.New("server-call-error"),
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().ListServices(gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewListServicesCommand)
		})
	}
}
