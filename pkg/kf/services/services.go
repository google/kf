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

package services

import (
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml options.go

// KfMarketplace contains information to describe the
// services and plans available in the catalog.
type KfMarketplace struct {
	Services []servicecatalog.Class
	Plans    []servicecatalog.Plan
}

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {

	// DeleteService removes an instance of a service on the cluster.
	DeleteService(instanceName string, opts ...DeleteServiceOption) error

	// GetService gets an instance of a service on the cluster.
	GetService(instanceName string, opts ...GetServiceOption) (*v1beta1.ServiceInstance, error)

	// ListServices lists instances of services on the cluster.
	ListServices(opts ...ListServicesOption) (*v1beta1.ServiceInstanceList, error)

	// Marketplace lists available services and plans in the marketplace.
	Marketplace(opts ...MarketplaceOption) (*KfMarketplace, error)

	// BrokerName fetches the service broker name for a service.
	BrokerName(service v1beta1.ServiceInstance, opts ...BrokerNameOption) (string, error)
}

// SClientFactory creates a Service Catalog client.
type SClientFactory func(namespace string) servicecatalog.SvcatClient

// NewClient creates a new client capable of interacting siwht service catalog
// services.
func NewClient(sclient SClientFactory) ClientInterface {
	return &Client{
		createSvcatClient: sclient,
	}
}

// Client is an implementation of ClientInterface that works with the Service Catalog.
type Client struct {
	createSvcatClient SClientFactory
}

// DeleteService removes an instance of a service on the cluster.
func (c *Client) DeleteService(instanceName string, opts ...DeleteServiceOption) error {
	cfg := DeleteServiceOptionDefaults().Extend(opts).toConfig()

	svcat := c.createSvcatClient(cfg.Namespace)
	return svcat.Deprovision(cfg.Namespace, instanceName)
}

// GetService gets an instance of a service on the cluster.
func (c *Client) GetService(instanceName string, opts ...GetServiceOption) (*v1beta1.ServiceInstance, error) {
	cfg := GetServiceOptionDefaults().Extend(opts).toConfig()

	svcat := c.createSvcatClient(cfg.Namespace)
	return svcat.RetrieveInstance(cfg.Namespace, instanceName)
}

// ListServices lists instances of services on the cluster.
func (c *Client) ListServices(opts ...ListServicesOption) (*v1beta1.ServiceInstanceList, error) {
	cfg := ListServicesOptionDefaults().Extend(opts).toConfig()

	svcat := c.createSvcatClient(cfg.Namespace)

	// RetrieveInstances(ns, classFilter, planFilter string)
	return svcat.RetrieveInstances(cfg.Namespace, "", "")
}

// Marketplace lists available services and plans in the marketplace.
func (c *Client) Marketplace(opts ...MarketplaceOption) (*KfMarketplace, error) {
	cfg := MarketplaceOptionDefaults().Extend(opts).toConfig()

	svcat := c.createSvcatClient(cfg.Namespace)

	scope := servicecatalog.ScopeOptions{
		Namespace: cfg.Namespace,
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
func (c *Client) BrokerName(service v1beta1.ServiceInstance, opts ...BrokerNameOption) (string, error) {
	cfg := BrokerNameOptionDefaults().Extend(opts).toConfig()
	svcat := c.createSvcatClient(cfg.Namespace)

	class, err := svcat.RetrieveClassByName(service.Spec.ClusterServiceClassExternalName, servicecatalog.ScopeOptions{
		Scope: servicecatalog.ClusterScope,
	})

	if err != nil {
		return "", err
	}

	return class.GetServiceBrokerName(), nil
}
