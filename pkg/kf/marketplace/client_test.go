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

package marketplace

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	fakescclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/poy/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
)

func TestClient_Marketplace(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName  string
		Namespace     string
		GetClassesErr error
		GetPlansErr   error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			Namespace:    "custom-namespace",
			ExpectErr:    nil,
		},
		"error in get classes": {
			InstanceName:  "instance-name",
			GetClassesErr: errors.New("server-err"),
			ExpectErr:     errors.New("server-err"),
		},
		"error in get plans": {
			InstanceName: "instance-name",
			GetPlansErr:  errors.New("server-err"),
			ExpectErr:    errors.New("server-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.RetrieveClassesStub = func(scope servicecatalog.ScopeOptions) ([]servicecatalog.Class, error) {
				testutil.AssertEqual(t, "namespace", tc.Namespace, scope.Namespace)

				return nil, tc.GetClassesErr
			}

			fakeClient.RetrievePlansStub = func(classFilter string, scope servicecatalog.ScopeOptions) ([]servicecatalog.Plan, error) {
				testutil.AssertEqual(t, "namespace", tc.Namespace, scope.Namespace)
				testutil.AssertEqual(t, "classFilter", "", classFilter)

				return nil, tc.GetPlansErr
			}

			client := NewClient(func(ns string) servicecatalog.SvcatClient {
				testutil.AssertEqual(t, "namespace", tc.Namespace, ns)

				return fakeClient
			}, nil)

			_, actualErr := client.Marketplace(tc.Namespace)
			if tc.ExpectErr != nil || actualErr != nil {
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

				return
			}

			testutil.AssertEqual(t, "calls to RetrieveClasses", 1, fakeClient.RetrieveClassesCallCount())
			testutil.AssertEqual(t, "calls to RetrievePlans", 1, fakeClient.RetrievePlansCallCount())
		})
	}
}

func TestClient_BrokerName(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		ExpectedName                   string
		Namespace                      string
		ExpectErr                      error
		ExpectedRetrieveClassByNameErr error
	}{
		"returns broker name": {
			ExpectedName: "some-broker-name",
		},
		"fetching class fails": {
			ExpectedRetrieveClassByNameErr: errors.New("some-error"),
			ExpectErr:                      errors.New("some-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}
			fakeClass := &fakeClass{brokerName: tc.ExpectedName}
			expectedSvc := v1beta1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tc.Namespace,
				},
			}

			fakeClient.RetrieveClassByNameStub = func(name string, opts servicecatalog.ScopeOptions) (servicecatalog.Class, error) {
				return fakeClass, tc.ExpectedRetrieveClassByNameErr
			}

			client := NewClient(func(ns string) servicecatalog.SvcatClient {
				testutil.AssertEqual(t, "namespace", tc.Namespace, ns)
				return fakeClient
			}, nil)

			name, actualErr := client.BrokerName(expectedSvc)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "broker name", tc.ExpectedName, name)
		})
	}
}

// fakeClass implements servicecatalog.Class. There isn't a fake provided.
type fakeClass struct {
	servicecatalog.Class
	brokerName string
}

func (f *fakeClass) GetServiceBrokerName() string {
	return f.brokerName
}

func TestClient_ListClusterPlans(t *testing.T) {
	plan := servicecatalogv1beta1.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-service-free",
		},
		Spec: servicecatalogv1beta1.ClusterServicePlanSpec{
			ClusterServiceBrokerName: "broker-a",
			CommonServicePlanSpec: servicecatalogv1beta1.CommonServicePlanSpec{
				ExternalName: "free",
				Free:         true,
			},
			ClusterServiceClassRef: servicecatalogv1beta1.ClusterObjectReference{
				Name: "db-service-id",
			},
		},
	}

	class := &servicecatalogv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-service-id",
		},
		Spec: servicecatalogv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: "broker-a",
			CommonServiceClassSpec: servicecatalogv1beta1.CommonServiceClassSpec{
				ExternalName: "db-service",
			},
		},
	}

	planList := &servicecatalogv1beta1.ClusterServicePlanList{
		Items: []servicecatalogv1beta1.ClusterServicePlan{
			plan,
		},
	}

	type args struct {
		filter ListPlanOptions
	}

	cases := map[string]struct {
		setup   func(t *testing.T) *fakescclient.Clientset
		args    args
		want    []servicecatalogv1beta1.ClusterServicePlan
		wantErr error
	}{
		"bad server call listing plans": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				client := fakescclient.NewSimpleClientset(planList, class)
				client.PrependReactor("list", "clusterserviceplans", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			args: args{
				filter: ListPlanOptions{},
			},
			wantErr: errors.New("server-call-error"),
		},
		"class lookup error": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				client := fakescclient.NewSimpleClientset(planList, class)
				client.PrependReactor("get", "clusterserviceclasses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("class-lookup-error")
				})
				return client
			},
			args: args{
				filter: ListPlanOptions{
					ServiceName: "force-lookup-to-match",
				},
			},
			wantErr: errors.New("class-lookup-error"),
		},
		"broker mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					BrokerName: "mismatch",
				},
			},
		},
		"plan mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					PlanName: "mismatch",
				},
			},
		},
		"service mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					ServiceName: "mismatch",
				},
			},
		},
		"matching": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					PlanName:    "free",
					ServiceName: "db-service",
					BrokerName:  "broker-a",
				},
			},
			want: []v1beta1.ClusterServicePlan{plan},
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			kclient := fakescclient.NewSimpleClientset()
			if tc.setup != nil {
				kclient = tc.setup(t)
			}
			c := &Client{
				kclient: kclient,
			}
			actualList, actualErr := c.ListClusterPlans(tc.args.filter)
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "plans", tc.want, actualList)
		})
	}
}

func TestClient_ListNamespacedPlans(t *testing.T) {
	plan := servicecatalogv1beta1.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-service-free",
			Namespace: "test-ns",
		},
		Spec: servicecatalogv1beta1.ServicePlanSpec{
			ServiceBrokerName: "broker-a",
			CommonServicePlanSpec: servicecatalogv1beta1.CommonServicePlanSpec{
				ExternalName: "free",
				Free:         true,
			},
			ServiceClassRef: servicecatalogv1beta1.LocalObjectReference{
				Name: "db-service-id",
			},
		},
	}

	class := &servicecatalogv1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-service-id",
			Namespace: "test-ns",
		},
		Spec: servicecatalogv1beta1.ServiceClassSpec{
			ServiceBrokerName: "broker-a",
			CommonServiceClassSpec: servicecatalogv1beta1.CommonServiceClassSpec{
				ExternalName: "db-service",
			},
		},
	}

	planList := &servicecatalogv1beta1.ServicePlanList{
		Items: []servicecatalogv1beta1.ServicePlan{
			plan,
		},
	}

	type args struct {
		namespace string
		filter    ListPlanOptions
	}

	cases := map[string]struct {
		setup   func(t *testing.T) *fakescclient.Clientset
		args    args
		want    []servicecatalogv1beta1.ServicePlan
		wantErr error
	}{
		"bad server call listing plans": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				client := fakescclient.NewSimpleClientset(planList, class)
				client.PrependReactor("list", "serviceplans", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			args: args{
				filter: ListPlanOptions{},
			},
			wantErr: errors.New("server-call-error"),
		},
		"class lookup error": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				client := fakescclient.NewSimpleClientset(planList, class)
				client.PrependReactor("get", "serviceclasses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("class-lookup-error")
				})
				return client
			},
			args: args{
				filter: ListPlanOptions{
					ServiceName: "force-lookup-to-match",
				},
			},
			wantErr: errors.New("class-lookup-error"),
		},
		"broker mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					BrokerName: "mismatch",
				},
			},
		},
		"plan mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					PlanName: "mismatch",
				},
			},
		},
		"service mismatch": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					ServiceName: "mismatch",
				},
			},
		},
		"matching": {
			setup: func(t *testing.T) *fakescclient.Clientset {
				return fakescclient.NewSimpleClientset(planList, class)
			},
			args: args{
				filter: ListPlanOptions{
					PlanName:    "free",
					ServiceName: "db-service",
					BrokerName:  "broker-a",
				},
			},
			want: []v1beta1.ServicePlan{plan},
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			kclient := fakescclient.NewSimpleClientset()
			if tc.setup != nil {
				kclient = tc.setup(t)
			}
			c := &Client{
				kclient: kclient,
			}
			actualList, actualErr := c.ListNamespacedPlans(tc.args.namespace, tc.args.filter)
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "plans", tc.want, actualList)
		})
	}
}
