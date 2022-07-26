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
	"context"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fakekfclient "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestClient_Marketplace(t *testing.T) {
	const namespace = "some-kf-ns"

	nsBroker := &v1alpha1.ServiceBroker{}
	nsBroker.Name = "ns-broker-a"
	nsBroker.Namespace = namespace
	nsBroker.Status = v1alpha1.CommonServiceBrokerStatus{
		Services: []v1alpha1.ServiceOffering{
			{
				DisplayName: "db-service-ns",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "free"},
				},
			},
			{
				DisplayName: "some-ns-class",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "some-ns-plan"},
				},
			},
		},
	}

	clusterBroker := &v1alpha1.ClusterServiceBroker{}
	clusterBroker.Name = "broker-a"
	clusterBroker.Status = v1alpha1.CommonServiceBrokerStatus{
		Services: []v1alpha1.ServiceOffering{
			{
				DisplayName: "db-service",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "free"},
				},
			},
			{
				DisplayName: "some-cluster-class",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "some-cluster-plan"},
				},
			},
		},
	}
	volumeBroker := &v1alpha1.ClusterServiceBroker{}
	volumeBroker.Name = "volume-broker-a"
	volumeBroker.Status = v1alpha1.CommonServiceBrokerStatus{
		Services: []v1alpha1.ServiceOffering{
			{
				DisplayName: "volume-class",
				Plans: []v1alpha1.ServicePlan{
					{DisplayName: "volume-plan"},
				},
			},
		},
	}
	kfClient := fakekfclient.NewSimpleClientset(
		nsBroker,
		clusterBroker,
		volumeBroker,
	).KfV1alpha1()

	client := NewClient(kfClient)
	marketplace, err := client.Marketplace(context.Background(), namespace)

	testutil.AssertNil(t, "marketplace error", err)

	classNames := sets.NewString()
	planNames := sets.NewString()

	marketplace.WalkServicePlans(func(l PlanLineage) {
		planNames.Insert(l.String())
		classNames.Insert(l.OfferingLineage.String())
	})

	testutil.AssertEqual(t, "plans", sets.NewString(
		"/broker-a/db-service/free",
		"/broker-a/some-cluster-class/some-cluster-plan",
		"/volume-broker-a/volume-class/volume-plan",
		"some-kf-ns/ns-broker-a/db-service-ns/free",
		"some-kf-ns/ns-broker-a/some-ns-class/some-ns-plan",
	), planNames)
	testutil.AssertEqual(t, "classes", sets.NewString(
		"/broker-a/db-service",
		"/broker-a/some-cluster-class",
		"/volume-broker-a/volume-class",
		"some-kf-ns/ns-broker-a/db-service-ns",
		"some-kf-ns/ns-broker-a/some-ns-class",
	), classNames)
}

func TestMarketplace_ListClusterPlans(t *testing.T) {
	t.Parallel()

	broker := &v1alpha1.ClusterServiceBroker{}
	broker.Name = "broker-a"
	broker.Namespace = "" // cluster
	broker.Status.Services = []v1alpha1.ServiceOffering{
		{
			DisplayName: "db-service",
			Plans: []v1alpha1.ServicePlan{
				{
					DisplayName: "free",
					Free:        true,
				},
				{
					DisplayName: "some-cluster-plan",
					Free:        true,
				},
			},
		},
	}

	fakeCatalog := &KfMarketplace{}
	fakeCatalog.Brokers = append(fakeCatalog.Brokers, broker)

	type args struct {
		filter ListPlanOptions
	}

	cases := map[string]struct {
		catalog *KfMarketplace
		args    args
		want    []string
	}{
		"broker mismatch": {
			catalog: fakeCatalog,
			args: args{
				filter: ListPlanOptions{
					BrokerName: "mismatch",
				},
			},
		},
		"plan mismatch": {
			catalog: fakeCatalog,
			args: args{
				filter: ListPlanOptions{
					PlanName: "mismatch",
				},
			},
		},
		"service mismatch": {
			catalog: fakeCatalog,
			args: args{
				filter: ListPlanOptions{
					ServiceName: "mismatch",
				},
			},
		},
		"matching": {
			catalog: fakeCatalog,
			args: args{
				filter: ListPlanOptions{
					PlanName:    "free",
					ServiceName: "db-service",
					BrokerName:  "broker-a",
				},
			},
			want: []string{
				"/broker-a/db-service/free",
			},
		},
		"multiple plans": {
			catalog: fakeCatalog,
			args: args{
				filter: ListPlanOptions{},
			},
			want: []string{
				"/broker-a/db-service/free",
				"/broker-a/db-service/some-cluster-plan",
			},
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualList := tc.catalog.ListClusterPlans(tc.args.filter)
			names := sets.NewString()
			for _, item := range actualList {
				names.Insert(item.String())
			}

			testutil.AssertEqual(t, "plans", tc.want, names.List())
		})
	}
}

func TestMarketplace_ListNamespacedPlans(t *testing.T) {
	t.Parallel()

	const namespace = "test-ns"

	broker := &v1alpha1.ServiceBroker{}
	broker.Name = "ns-broker-a"
	broker.Namespace = namespace
	broker.Status.Services = []v1alpha1.ServiceOffering{
		{
			DisplayName: "db-service-ns",
			Plans: []v1alpha1.ServicePlan{
				{
					DisplayName: "free",
					Free:        true,
				},
				{
					DisplayName: "some-plan",
					Free:        true,
				},
			},
		},
	}

	fakeCatalog := &KfMarketplace{}
	fakeCatalog.Brokers = append(fakeCatalog.Brokers, broker)

	type args struct {
		namespace string
		filter    ListPlanOptions
	}

	cases := map[string]struct {
		catalog *KfMarketplace
		args    args
		want    []string
	}{
		"broker mismatch": {
			catalog: fakeCatalog,
			args: args{
				namespace: namespace,
				filter: ListPlanOptions{
					BrokerName: "mismatch",
				},
			},
		},
		"plan mismatch": {
			catalog: fakeCatalog,
			args: args{
				namespace: namespace,
				filter: ListPlanOptions{
					PlanName: "mismatch",
				},
			},
		},
		"service mismatch": {
			catalog: fakeCatalog,
			args: args{
				namespace: namespace,
				filter: ListPlanOptions{
					ServiceName: "mismatch",
				},
			},
		},
		"matching": {
			catalog: fakeCatalog,
			args: args{
				namespace: namespace,
				filter: ListPlanOptions{
					PlanName:    "free",
					ServiceName: "db-service-ns",
					BrokerName:  "ns-broker-a",
				},
			},
			want: []string{
				"test-ns/ns-broker-a/db-service-ns/free",
			},
		},
		"multiple plans": {
			catalog: fakeCatalog,
			args: args{
				namespace: namespace,
				filter:    ListPlanOptions{},
			},
			want: []string{
				"test-ns/ns-broker-a/db-service-ns/free",
				"test-ns/ns-broker-a/db-service-ns/some-plan",
			},
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualList := tc.catalog.ListNamespacedPlans(tc.args.namespace, tc.args.filter)
			names := sets.NewString()
			for _, item := range actualList {
				names.Insert(item.String())
			}

			testutil.AssertEqual(t, "plans", tc.want, names.List())
		})
	}
}
