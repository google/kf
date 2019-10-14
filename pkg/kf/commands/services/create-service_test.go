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
	"github.com/google/kf/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	marketplacefake "github.com/google/kf/pkg/kf/marketplace/fake"
	servicesfake "github.com/google/kf/pkg/kf/services/fake"
	"github.com/google/kf/pkg/kf/testutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewCreateServiceCommand(t *testing.T) {
	type fakes struct {
		services    *servicesfake.FakeClient
		marketplace *marketplacefake.FakeClientInterface
	}

	cases := map[string]struct {
		args      []string
		namespace string
		setup     func(*testing.T, fakes)
		expectErr error
	}{
		// user errors
		"bad number of args": {
			expectErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"bad namespace": {
			args:      []string{"db-service", "free", "mydb"},
			expectErr: errors.New(utils.EmptyNamespaceError),
		},
		"bad path": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "--config=/some/bad/path"},
			expectErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},

		// server errors
		"cluster plan failure": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any()).Return(nil, errors.New("cluster-list-failure"))
			},
			expectErr: errors.New("cluster-list-failure"),
		},
		"namespace plan failure": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any())
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any()).Return(nil, errors.New("ns-list-failure"))
			},
			expectErr: errors.New("ns-list-failure"),
		},

		// plans errors
		"no plans all brokers": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any())
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any())
			},
			expectErr: errors.New("no plan free found for class db-service for all service-brokers"),
		},
		"no plans specific broker": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "-b", "testbroker"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any())
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any())
			},
			expectErr: errors.New("no plan free found for class db-service for the service-broker testbroker"),
		},
		"multiple plan matches": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "-b", "testbroker"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any()).Return([]servicecatalogv1beta1.ClusterServicePlan{{}}, nil)
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any()).Return([]servicecatalogv1beta1.ServicePlan{{}}, nil)
			},
			expectErr: errors.New("plans matched from multiple brokers, specify a broker with --broker"),
		},

		// good results
		"cluster": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "-b", "testbroker"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any()).Return([]servicecatalogv1beta1.ClusterServicePlan{{}}, nil)
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any())

				fakes.services.EXPECT().Create("test-ns", &servicecatalogv1beta1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: "test-ns",
					},
					Spec: servicecatalogv1beta1.ServiceInstanceSpec{
						PlanReference: servicecatalogv1beta1.PlanReference{
							ClusterServicePlanExternalName:  "free",
							ClusterServiceClassExternalName: "db-service",
						},
						Parameters: &runtime.RawExtension{
							Raw: json.RawMessage(`{}`),
						},
					},
				})

				fakes.services.EXPECT().WaitForProvisionSuccess(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},
		"namespaced": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "-b", "testbroker"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any())
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any()).Return([]servicecatalogv1beta1.ServicePlan{{}}, nil)

				fakes.services.EXPECT().Create("test-ns", &servicecatalogv1beta1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: "test-ns",
					},
					Spec: servicecatalogv1beta1.ServiceInstanceSpec{
						PlanReference: servicecatalogv1beta1.PlanReference{
							ServicePlanExternalName:  "free",
							ServiceClassExternalName: "db-service",
						},
						Parameters: &runtime.RawExtension{
							Raw: json.RawMessage(`{}`),
						},
					},
				})

				fakes.services.EXPECT().WaitForProvisionSuccess(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},
		"async": {
			namespace: "test-ns",
			args:      []string{"db-service", "free", "mydb", "--async"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.marketplace.EXPECT().ListClusterPlans(gomock.Any())
				fakes.marketplace.EXPECT().ListNamespacedPlans(gomock.Any(), gomock.Any()).Return([]servicecatalogv1beta1.ServicePlan{{}}, nil)
				fakes.services.EXPECT().Create(gomock.Any(), gomock.Any())
				// expect WaitForConditionReadyTrue not to be called
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sClient := servicesfake.NewFakeClient(ctrl)
			mClient := marketplacefake.NewFakeClientInterface(ctrl)
			if tc.setup != nil {
				tc.setup(t, fakes{
					services:    sClient,
					marketplace: mClient,
				})
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.namespace,
			}

			cmd := servicescmd.NewCreateServiceCommand(p, sClient, mClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			_, actualErr := cmd.ExecuteC()
			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			}
		})
	}
}
