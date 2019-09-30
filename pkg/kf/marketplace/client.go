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
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KfMarketplace contains information to describe the
// services and plans available in the catalog.
type KfMarketplace struct {
	Services []servicecatalog.Class
	Plans    []servicecatalog.Plan
}

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// Marketplace lists available services and plans in the marketplace.
	Marketplace(namespace string) (*KfMarketplace, error)

	// BrokerName fetches the service broker name for a service.
	BrokerName(service v1beta1.ServiceInstance) (string, error)
}

// SClientFactory creates a Service Catalog client.
type SClientFactory func(namespace string) servicecatalog.SvcatClient

// NewClient creates a new client capable of interacting siwht service catalog
// services.
func NewClient(sclient SClientFactory, kclient servicecatalogclient.Interface) ClientInterface {
	return &Client{
		createSvcatClient: sclient,
		kclient:           kclient,
	}
}

// Client is an implementation of ClientInterface that works with the Service Catalog.
type Client struct {
	createSvcatClient SClientFactory
	kclient           servicecatalogclient.Interface
}

// Marketplace lists available services and plans in the marketplace.
func (c *Client) Marketplace(namespace string) (*KfMarketplace, error) {
	svcat := c.createSvcatClient(namespace)

	scope := servicecatalog.ScopeOptions{
		Namespace: namespace,
		Scope:     servicecatalog.AllScope,
	}

	classes, err := svcat.RetrieveClasses(scope)
	if err != nil {
		return nil, err
	}

	// an empty first param gets all plans
	plans, err := svcat.RetrievePlans("", scope)
	if err != nil {
		return nil, err
	}

	return &KfMarketplace{
		Services: classes,
		Plans:    plans,
	}, nil
}

// BrokerName fetches the service broker name for a service.
func (c *Client) BrokerName(service v1beta1.ServiceInstance) (string, error) {
	svcat := c.createSvcatClient(service.GetNamespace())

	scope := servicecatalog.ScopeOptions{
		Scope: servicecatalog.ClusterScope,
	}
	className := service.Spec.ClusterServiceClassExternalName

	if service.Spec.ServiceClassRef != nil {
		scope = servicecatalog.ScopeOptions{
			Namespace: service.GetNamespace(),
			Scope:     servicecatalog.NamespaceScope,
		}
		className = service.Spec.ServiceClassExternalName
	}

	class, err := svcat.RetrieveClassByName(className, scope)

	if err != nil {
		return "", err
	}

	return class.GetServiceBrokerName(), nil
}

// ListPlanOptions holds additional filtering options used when listing plans.
type ListPlanOptions struct {
	PlanName    string
	ServiceName string
	BrokerName  string
}

func (c *Client) ListClusterPlans(filter ListPlanOptions) ([]servicecatalogv1beta1.ClusterServicePlan, error) {
	var matchingPlans []servicecatalogv1beta1.ClusterServicePlan

	plans, err := c.kclient.ServicecatalogV1beta1().
		ClusterServicePlans().
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, plan := range plans.Items {
		if filter.PlanName != "" && filter.PlanName != plan.GetExternalName() {
			continue
		}

		if filter.BrokerName != "" && filter.BrokerName != plan.Spec.ClusterServiceBrokerName {
			continue
		}

		if filter.ServiceName != "" {
			class, err := c.kclient.ServicecatalogV1beta1().
				ClusterServiceClasses().
				Get(plan.Spec.ClusterServiceClassRef.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			if filter.ServiceName != class.Spec.ExternalName {
				continue
			}
		}

		matchingPlans = append(matchingPlans, plan)
	}

	return matchingPlans, nil
}

func (c *Client) ListNamespacedPlans(namespace string, filter ListPlanOptions) ([]servicecatalogv1beta1.ServicePlan, error) {
	var matchingPlans []servicecatalogv1beta1.ServicePlan

	plans, err := c.kclient.ServicecatalogV1beta1().
		ServicePlans(namespace).
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, plan := range plans.Items {
		if filter.PlanName != "" && filter.PlanName != plan.GetExternalName() {
			continue
		}

		if filter.BrokerName != "" && filter.BrokerName != plan.Spec.ServiceBrokerName {
			continue
		}

		if filter.ServiceName != "" {
			class, err := c.kclient.ServicecatalogV1beta1().
				ServiceClasses(namespace).
				Get(plan.Spec.ServiceClassRef.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			if filter.ServiceName != class.Spec.ExternalName {
				continue
			}
		}

		matchingPlans = append(matchingPlans, plan)
	}

	return matchingPlans, nil
}
