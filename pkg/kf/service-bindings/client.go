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

package servicebindings

import (
	"fmt"

	"github.com/google/kf/pkg/kf/secrets"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	clientv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml options.go

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

	// GetVcapServices gets a VCAP_SERVICES compatible environment variable.
	GetVcapServices(appName string, opts ...GetVcapServicesOption) (VcapServicesMap, error)
}

// NewClient creates a new client capable of interacting with service catalog
// services.
func NewClient(c clientv1beta1.ServicecatalogV1beta1Interface, sc secrets.ClientInterface) ClientInterface {
	return &Client{
		c:  c,
		sc: sc,
	}
}

type Client struct {
	c  clientv1beta1.ServicecatalogV1beta1Interface
	sc secrets.ClientInterface
}

// Create binds a service instance to an app.
func (c *Client) Create(serviceInstanceName, appName string, opts ...CreateOption) (*apiv1beta1.ServiceBinding, error) {
	cfg := CreateOptionDefaults().Extend(opts).toConfig()

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

	return c.c.ServiceBindings(cfg.Namespace).Create(request)
}

// Delete unbinds a service instance from an app.
func (c *Client) Delete(serviceInstanceName, appName string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	bindingReference := serviceBindingName(appName, serviceInstanceName)
	return c.c.ServiceBindings(cfg.Namespace).Delete(bindingReference, &v1.DeleteOptions{})
}

// List queries Kubernetes for service bindings.
func (c *Client) List(opts ...ListOption) ([]apiv1beta1.ServiceBinding, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	bindings, err := c.c.ServiceBindings(cfg.Namespace).List(v1.ListOptions{})
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

// GetVcapServices gets a VCAP_SERVICES compatible environment variable.
func (c *Client) GetVcapServices(appName string, opts ...GetVcapServicesOption) (VcapServicesMap, error) {
	cfg := GetVcapServicesOptionDefaults().Extend(opts).toConfig()

	bindings, err := c.List(WithListAppName(appName), WithListNamespace(cfg.Namespace))
	if err != nil {
		return nil, err
	}

	out := VcapServicesMap{}
	for _, binding := range bindings {
		instance, err := c.c.ServiceInstances(cfg.Namespace).Get(binding.Spec.InstanceRef.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("couldn't create VCAP_SERVICES, couldn't get instance for binding %s: %v", binding.Name, err)
		}

		secret, err := c.sc.Get(binding.Spec.SecretName, secrets.WithGetNamespace(cfg.Namespace))
		if err != nil {
			if cfg.FailOnBadSecret {
				return nil, fmt.Errorf("couldn't create VCAP_SERVICES, the secret for binding %s couldn't be fetched: %v", binding.Name, err)
			} else {
				continue
			}
		}

		out.Add(NewVcapService(*instance, binding, secret))
	}

	return out, nil
}

// serviceBindingName is the primary key for service bindings consisting of the
// app name paired with the instance name to duplicate CF's 1:1 binding limit.
func serviceBindingName(appName, instanceName string) string {
	return fmt.Sprintf("kf-binding-%s-%s", appName, instanceName)
}
