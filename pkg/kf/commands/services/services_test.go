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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	fakeapps "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	fakemarketplace "github.com/google/kf/pkg/kf/marketplace/fake"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/services/fake"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewServicesCommand(t *testing.T) {
	cases := map[string]struct {
		serviceTest
		AppSetup         func(t *testing.T, f *fakeapps.FakeClient)
		MarketplaceSetup func(t *testing.T, f *fakemarketplace.FakeClientInterface)
	}{
		"too many params": {
			serviceTest: serviceTest{
				Args:        []string{"foo", "bar"},
				ExpectedErr: errors.New("accepts 0 arg(s), received 2"),
			},
		},
		"custom namespace": {
			serviceTest: serviceTest{
				Namespace: "test-ns",
				Setup: func(t *testing.T, f *fake.FakeClient) {
					f.EXPECT().List("test-ns").Return([]v1beta1.ServiceInstance{}, nil)
				},
			},
		},
		"empty namespace": {
			serviceTest: serviceTest{
				ExpectedErr: errors.New(utils.EmptyNamespaceError),
			},
		},
		"empty result": {
			serviceTest: serviceTest{
				Namespace: "test-ns",
				Setup: func(t *testing.T, f *fake.FakeClient) {
					f.EXPECT().List(gomock.Any()).Return([]v1beta1.ServiceInstance{}, nil)
				},
				ExpectedErr: nil, // explicitly expecting no failure with zero length list
			},
		},
		"fetching apps fails": {
			AppSetup: func(t *testing.T, f *fakeapps.FakeClient) {
				f.EXPECT().List("test-ns").Return(nil, errors.New("some-error"))
			},
			serviceTest: serviceTest{
				Namespace: "test-ns",
				Setup: func(t *testing.T, f *fake.FakeClient) {
					f.EXPECT().List(gomock.Any()).Return([]v1beta1.ServiceInstance{}, nil)
				},
				ExpectedErr: errors.New("some-error"),
			},
		},
		"fetching broker name fails": {
			MarketplaceSetup: func(t *testing.T, f *fakemarketplace.FakeClientInterface) {
				f.EXPECT().BrokerName(gomock.Any()).Return("", errors.New("fetch-broker-name-error")).Times(2)
			},
			serviceTest: serviceTest{
				Namespace: "test-ns",
				Setup: func(t *testing.T, f *fake.FakeClient) {
					serviceList := []v1beta1.ServiceInstance{
						*dummyServerInstance("service-1"),
						*dummyServerInstance("service-2"),
					}
					f.EXPECT().List(gomock.Any()).Return(serviceList, nil)
				},
				ExpectedErr: errors.New("fetch-broker-name-error"),
				ExpectedStrings: []string{
					"service-1", "service-2", // service instances still displayed with error msg
					"fetch-broker-name-error",
				},
			},
		},
		"full result": {
			AppSetup: func(t *testing.T, f *fakeapps.FakeClient) {
				f.EXPECT().List("test-ns").Return([]v1alpha1.App{
					boundApp("app-1", "service-1"),
					boundApp("app-2", "service-2"),
				}, nil)
			},
			serviceTest: serviceTest{
				Namespace: "test-ns",
				Setup: func(t *testing.T, f *fake.FakeClient) {
					// We'll take the conditions off this so it has to show an
					// unknown state
					service1 := *dummyServerInstance("service-1")
					service1.Status.Conditions = nil

					serviceList := []v1beta1.ServiceInstance{
						service1,
						*dummyServerInstance("service-2"),
					}
					f.EXPECT().List(gomock.Any()).Return(serviceList, nil)
				},
				ExpectedStrings: []string{
					"service-1", "service-2", // Binding Names
					"app-1", "app-2", // Bound Apps
					"some-broker",              // Broker Names
					"CorrectStatus", "Unknown", // Last Operation
				},
			},
		},
		"bad server call": {
			serviceTest: serviceTest{
				Namespace:   "test-ns",
				ExpectedErr: errors.New("server-call-error"),
				Setup: func(t *testing.T, f *fake.FakeClient) {
					f.EXPECT().List(gomock.Any()).Return(nil, errors.New("server-call-error"))
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			appClient := fakeapps.NewFakeClient(gomock.NewController(t))
			if tc.AppSetup != nil {
				tc.AppSetup(t, appClient)
			} else {
				// Give default empty app response
				appClient.EXPECT().List(gomock.Any())
			}

			marketplaceClient := fakemarketplace.NewFakeClientInterface(gomock.NewController(t))
			if tc.MarketplaceSetup != nil {
				tc.MarketplaceSetup(t, marketplaceClient)
			} else {
				// Give default empty broker response
				marketplaceClient.EXPECT().BrokerName(gomock.Any()).Return("some-broker", nil).AnyTimes()
			}

			runTest(t, tc.serviceTest, func(p *config.KfParams, client services.Client) *cobra.Command {
				return servicescmd.NewListServicesCommand(p, client, appClient, marketplaceClient)
			})
		})
	}
}

func boundApp(name, bindingName string) v1alpha1.App {
	return v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.AppSpec{
			ServiceBindings: []v1alpha1.AppSpecServiceBinding{
				{BindingName: bindingName},
			},
		},
	}
}
