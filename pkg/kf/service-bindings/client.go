package servicebindings

import (
	"fmt"

	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	clientv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml

const (
	// BindingNameLabel is the label used on bindings to define what VCAP name the secret should be rooted under.
	BindingNameLabel = "kf-binding-name"
	// AppNameLabel is the label used on bindings to define which app the binding belongs to.
	AppNameLabel = "kf-app-name"
)

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// Create binds a service instance to an app.
	Create(serviceInstanceName, appName string, opts ...CreateOption) (*apiv1beta1.ServiceBinding, error)

	// Delete removes a service binding from an app.
	Delete(serviceInstanceName, appName string, opts ...DeleteOption) error

	// List queries Kubernetes for service bindings.
	List(opts ...ListOption) ([]apiv1beta1.ServiceBinding, error)
}

// SClientFactory creates a Service Catalog client.
type SClientFactory func() (clientv1beta1.ServicecatalogV1beta1Interface, error)

// NewClient creates a new client capable of interacting with service catalog
// services.
func NewClient(sclient SClientFactory) ClientInterface {
	return &Client{
		createSvcatClient: sclient,
	}
}

type Client struct {
	createSvcatClient SClientFactory
}

// Create binds a service instance to an app.
func (c *Client) Create(serviceInstanceName, appName string, opts ...CreateOption) (*apiv1beta1.ServiceBinding, error) {
	cfg := CreateOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient()
	if err != nil {
		return nil, err
	}

	bindingName := cfg.BindingName
	if bindingName == "" {
		bindingName = serviceInstanceName
	}

	bindingReference := serviceBindingName(appName, serviceInstanceName)
	request := &apiv1beta1.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      bindingReference,
			Namespace: cfg.Namespace,
			Labels: map[string]string{
				BindingNameLabel: bindingName,
				AppNameLabel:     appName,
			},
		},
		Spec: apiv1beta1.ServiceBindingSpec{
			InstanceRef: apiv1beta1.LocalObjectReference{
				Name: serviceInstanceName,
			},
			SecretName: bindingReference,
			Parameters: servicecatalog.BuildParameters(cfg.Params),
		},
	}

	return svcat.ServiceBindings(cfg.Namespace).Create(request)
}

// Delete unbinds a service instance from an app.
func (c *Client) Delete(serviceInstanceName, appName string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient()
	if err != nil {
		return err
	}

	bindingReference := serviceBindingName(appName, serviceInstanceName)
	return svcat.ServiceBindings(cfg.Namespace).Delete(bindingReference, &v1.DeleteOptions{})
}

// List queries Kubernetes for service bindings.
func (c *Client) List(opts ...ListOption) ([]apiv1beta1.ServiceBinding, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	svcat, err := c.createSvcatClient()
	if err != nil {
		return nil, err
	}
	bindings, err := svcat.ServiceBindings(cfg.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Filter the results
	filterByServiceInstance := cfg.ServiceInstance != ""
	filterByAppName := cfg.AppName != ""

	var filtered []apiv1beta1.ServiceBinding
	for _, binding := range bindings.Items {
		if filterByServiceInstance && binding.Spec.InstanceRef.Name != cfg.ServiceInstance {
			continue
		}

		// NOTE: this _could_ be done with a label selector, but we'll do it here
		// to reduce the cognitive overhead of filtering in multiple locations.
		if filterByAppName && binding.Labels[AppNameLabel] != cfg.AppName {
			continue
		}

		filtered = append(filtered, binding)
	}

	return filtered, nil
}

// serviceBindingName is the primary key for service bindings consisting of the
// app name paired with the instance name to duplicate CF's 1:1 binding limit.
func serviceBindingName(appName, instanceName string) string {
	return fmt.Sprintf("kf-binding-%s-%s", appName, instanceName)
}
