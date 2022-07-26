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
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclientv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KfMarketplace contains information to describe the
// services and plans available in the catalog.
type KfMarketplace struct {
	Brokers []v1alpha1.CommonServiceBroker
}

// OfferingLineage holds a broker/offering tuple where the broker is the parent
// of the offering.
type OfferingLineage struct {
	Broker          v1alpha1.CommonServiceBroker
	ServiceOffering v1alpha1.ServiceOffering
}

// String implements fmt.Stringer.
func (o *OfferingLineage) String() string {
	return strings.Join([]string{
		o.Broker.GetNamespace(),
		o.Broker.GetName(),
		o.ServiceOffering.DisplayName,
	}, "/")
}

// PlanLineage holds a broker/offering/plan tuple where the broker is the parent
// of the offering which is the parent of the plan.
type PlanLineage struct {
	OfferingLineage
	ServicePlan v1alpha1.ServicePlan
}

// String implements fmt.Stringer.
func (o *PlanLineage) String() string {
	return strings.Join([]string{
		o.OfferingLineage.String(),
		o.ServicePlan.DisplayName,
	}, "/")
}

// WalkServiceOfferings iterates through each broker/service tuple.
func (m *KfMarketplace) WalkServiceOfferings(callback func(OfferingLineage)) {
	for _, broker := range m.Brokers {
		for _, offering := range broker.GetServiceOfferings() {
			callback(OfferingLineage{
				Broker:          broker,
				ServiceOffering: offering,
			})
		}
	}
}

// WalkServicePlans iterates through each broker/service/plan tuple.
func (m *KfMarketplace) WalkServicePlans(callback func(PlanLineage)) {
	m.WalkServiceOfferings(func(ol OfferingLineage) {
		for _, plan := range ol.ServiceOffering.Plans {
			callback(PlanLineage{
				OfferingLineage: ol,
				ServicePlan:     plan,
			})
		}
	})
}

// ListClusterPlans gets cluster-wide plans matching the given filter.
func (m *KfMarketplace) ListClusterPlans(filter ListPlanOptions) []PlanLineage {
	return m.ListNamespacedPlans("", filter)
}

// ListNamespacedPlans gets namespaced plans matching the given filter.
func (m *KfMarketplace) ListNamespacedPlans(namespace string, filter ListPlanOptions) []PlanLineage {
	var out []PlanLineage

	m.WalkServicePlans(func(lineage PlanLineage) {
		if lineage.Broker.GetNamespace() != namespace {
			return
		}

		if filter.BrokerName != "" && lineage.Broker.GetName() != filter.BrokerName {
			return
		}

		if filter.ServiceName != "" && lineage.ServiceOffering.DisplayName != filter.ServiceName {
			return
		}

		if filter.PlanName != "" && lineage.ServicePlan.DisplayName != filter.PlanName {
			return
		}

		out = append(out, lineage)
	})

	return out
}

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// Marketplace lists available services and plans in the Kf OSB marketplace.
	Marketplace(ctx context.Context, namespace string) (*KfMarketplace, error)
}

// NewClient creates a new client capable of interacting with service catalog
// services.
func NewClient(kfclient kfclientv1alpha1.KfV1alpha1Interface) ClientInterface {
	return &Client{
		kfclient: kfclient,
	}
}

// Client is an implementation of ClientInterface that works with the Service Catalog.
type Client struct {
	kfclient kfclientv1alpha1.KfV1alpha1Interface
}

// Marketplace lists available services and plans in the Kf OSB marketplace.
func (c *Client) Marketplace(ctx context.Context, namespace string) (*KfMarketplace, error) {
	var out KfMarketplace

	// namespace scoped
	{
		brokers, err := c.kfclient.
			ServiceBrokers(namespace).
			List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for i := range brokers.Items {
			out.Brokers = append(out.Brokers, &brokers.Items[i])
		}
	}

	// cluster scoped
	{
		brokers, err := c.kfclient.
			ClusterServiceBrokers().
			List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for i := range brokers.Items {
			out.Brokers = append(out.Brokers, &brokers.Items[i])
		}
	}

	return &out, nil
}

// ListPlanOptions holds additional filtering options used when listing plans.
type ListPlanOptions struct {
	PlanName    string
	ServiceName string
	BrokerName  string
}
