package services

import (
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml

// KfMarketplace contains information to describe the
// services and plans available in the catalog.
type KfMarketplace struct {
	Services []servicecatalog.Class
	Plans    []servicecatalog.Plan
}

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// CreateService creates a new instance of a service on the cluster.
	CreateService(instanceName, serviceName, planName string, opts ...CreateServiceOption) (*v1beta1.ServiceInstance, error)

	// DeleteService removes an instance of a service on the cluster.
	DeleteService(instanceName string, opts ...DeleteServiceOption) error

	// GetService gets an instance of a service on the cluster.
	GetService(instanceName string, opts ...GetServiceOption) (*v1beta1.ServiceInstance, error)

	// ListServices lists instances of services on the cluster.
	ListServices(opts ...ListServicesOption) (*v1beta1.ServiceInstanceList, error)

	// Marketplace lists available services and plans in the marketplace.
	Marketplace(opts ...MarketplaceOption) (*KfMarketplace, error)
}

// SClientFactory creates a Service Catalog client.
type SClientFactory func(namespace string) (servicecatalog.SvcatClient, error)

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

// CreateService creates a new instance of a service on the cluster.
func (c *Client) CreateService(instanceName, serviceName, planName string, opts ...CreateServiceOption) (*v1beta1.ServiceInstance, error) {
	cfg := CreateServiceOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	// Provision(instanceName, className, planName string, opts *ProvisionOptions) (*v1beta1.ServiceInstance, error)
	return svcat.Provision(instanceName, serviceName, planName, &servicecatalog.ProvisionOptions{
		Namespace: cfg.Namespace,
		Params:    cfg.Params,
	})
}

// DeleteService removes an instance of a service on the cluster.
func (c *Client) DeleteService(instanceName string, opts ...DeleteServiceOption) error {
	cfg := DeleteServiceOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient(cfg.Namespace)
	if err != nil {
		return err
	}

	return svcat.Deprovision(cfg.Namespace, instanceName)
}

// GetService gets an instance of a service on the cluster.
func (c *Client) GetService(instanceName string, opts ...GetServiceOption) (*v1beta1.ServiceInstance, error) {
	cfg := GetServiceOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	return svcat.RetrieveInstance(cfg.Namespace, instanceName)
}

// ListServices lists instances of services on the cluster.
func (c *Client) ListServices(opts ...ListServicesOption) (*v1beta1.ServiceInstanceList, error) {
	cfg := ListServicesOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	// RetrieveInstances(ns, classFilter, planFilter string)
	return svcat.RetrieveInstances(cfg.Namespace, "", "")
}

// Marketplace lists available services and plans in the marketplace.
func (c *Client) Marketplace(opts ...MarketplaceOption) (*KfMarketplace, error) {
	cfg := MarketplaceOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient(cfg.Namespace)
	if err != nil {
		return nil, err
	}

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
