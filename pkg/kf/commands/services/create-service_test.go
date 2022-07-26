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
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/v2/pkg/kf/commands/services"
	"github.com/google/kf/v2/pkg/kf/marketplace"
	marketplacefake "github.com/google/kf/v2/pkg/kf/marketplace/fake"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	serviceinstancesfake "github.com/google/kf/v2/pkg/kf/serviceinstances/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewCreateServiceCommand(t *testing.T) {
	type fakes struct {
		services    *serviceinstancesfake.FakeClient
		secrets     *secretsfake.FakeClient
		marketplace *marketplacefake.FakeClientInterface
	}

	const mockNs = "test-ns"
	const mockServiceUID = "00000000-0000-0000-0000-000000000000"
	const mockPlanUID = "11111111-1111-1111-1111-111111111111"

	mockNsBroker := &v1alpha1.ServiceBroker{}
	mockNsBroker.Name = "namespaced-broker"
	mockNsBroker.Namespace = mockNs
	mockNsBroker.Status.Services = []v1alpha1.ServiceOffering{
		{
			DisplayName: "db-service",
			UID:         mockServiceUID,
			Tags:        []string{"ns", "db"},
			Plans: []v1alpha1.ServicePlan{
				{DisplayName: "free", UID: mockPlanUID},
			},
		},
	}

	mockClusterBroker := &v1alpha1.ClusterServiceBroker{}
	mockClusterBroker.Name = "cluster-broker"
	mockClusterBroker.Status.Services = []v1alpha1.ServiceOffering{
		{
			DisplayName: "db-service",
			UID:         mockServiceUID,
			Tags:        []string{"cluster", "db"},
			Plans: []v1alpha1.ServicePlan{
				{DisplayName: "free", UID: mockPlanUID},
			},
		},
	}

	mockMarketplace := &marketplace.KfMarketplace{
		Brokers: []v1alpha1.CommonServiceBroker{
			mockClusterBroker,
			mockNsBroker,
		},
	}

	cases := map[string]struct {
		args      []string
		namespace string
		enableOSB bool
		setup     func(*testing.T, fakes)
		expectErr error
	}{
		// user errors
		"bad number of args": {
			expectErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"bad namespace": {
			args:      []string{"db-service", "free", "mydb"},
			expectErr: errors.New(config.EmptySpaceError),
		},
		"bad path": {
			namespace: mockNs,
			args:      []string{"db-service", "free", "mydb", "-c=/some/bad/path"},
			expectErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},

		// server errors
		"marketplace failure": {
			namespace: mockNs,
			args:      []string{"db-service", "free", "mydb"},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(nil, errors.New("marketplace-failure"))
			},
			expectErr: errors.New("marketplace-failure"),
		},

		// plans errors
		"no plans all brokers": {
			namespace: mockNs,
			args:      []string{"db-service", "free", "mydb"},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(&marketplace.KfMarketplace{}, nil)
			},
			expectErr: errors.New("no plan free found for class db-service for all service-brokers"),
		},
		"no plans specific broker": {
			namespace: mockNs,
			args:      []string{"db-service", "badplan", "mydb", "-b", mockNsBroker.Name},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(&marketplace.KfMarketplace{}, nil)
			},
			expectErr: errors.New("no plan badplan found for class db-service for the service-broker namespaced-broker"),
		},
		"multiple plan matches": {
			namespace: mockNs,
			args:      []string{"db-service", "free", "mydb"},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)
			},
			expectErr: errors.New("plans matched from multiple brokers, specify a broker with --broker"),
		},

		// good results
		"cluster": {
			namespace: mockNs,
			args:      []string{"db-service", "free", "mydb", "-b", mockClusterBroker.Name, "--timeout", "600s"},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)

				fakes.services.EXPECT().Create(gomock.Any(), mockNs, &v1alpha1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: mockNs,
					},
					Spec: v1alpha1.ServiceInstanceSpec{
						ServiceType: v1alpha1.ServiceType{
							OSB: &v1alpha1.OSBInstance{
								BrokerName:              mockClusterBroker.Name,
								ClassName:               "db-service",
								ClassUID:                mockServiceUID,
								PlanName:                "free",
								PlanUID:                 mockPlanUID,
								Namespaced:              false,
								ProgressDeadlineSeconds: 600,
							},
						},
						ParametersFrom: corev1.LocalObjectReference{
							Name: v1alpha1.GenerateName("serviceinstance", "mydb", "params"),
						},
						Tags: []string{"cluster", "db"},
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage("{}"))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), mockNs, "mydb", gomock.Any())
			},
		},
		"namespaced": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "-b", mockNsBroker.Name},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)

				secretName := v1alpha1.GenerateName("serviceinstance", "mydb", "params")
				fakes.services.EXPECT().Create(gomock.Any(), mockNs, &v1alpha1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: mockNs,
					},
					Spec: v1alpha1.ServiceInstanceSpec{
						ServiceType: v1alpha1.ServiceType{
							OSB: &v1alpha1.OSBInstance{
								BrokerName:              mockNsBroker.Name,
								ClassName:               "db-service",
								ClassUID:                mockServiceUID,
								PlanName:                "free",
								PlanUID:                 mockPlanUID,
								Namespaced:              true,
								ProgressDeadlineSeconds: 1800,
							},
						},
						ParametersFrom: corev1.LocalObjectReference{
							Name: secretName,
						},
						Tags: []string{"db", "ns"},
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), secretName, json.RawMessage("{}"))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), mockNs, "mydb", gomock.Any())
			},
		},
		"async": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "--async", "--broker", mockNsBroker.Name},
			enableOSB: true,
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().Marketplace(gomock.Any(), gomock.Any()).Return(mockMarketplace, nil)
				fakes.services.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				// expect WaitForConditionReadyTrue not to be called
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sClient := serviceinstancesfake.NewFakeClient(ctrl)
			mClient := marketplacefake.NewFakeClientInterface(ctrl)
			secretClient := secretsfake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakes{
					services:    sClient,
					marketplace: mClient,
					secrets:     secretClient,
				})
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.namespace,
			}

			cmd := servicescmd.NewCreateServiceCommand(p, sClient, secretClient, mClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			_, actualErr := cmd.ExecuteC()
			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			}
		})
	}
}
