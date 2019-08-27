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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	"github.com/google/kf/pkg/kf/apps"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml options.go

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// Create binds a service instance to an app.
	Create(serviceInstanceName, appName string, opts ...CreateOption) (*v1alpha1.AppSpecServiceBinding, error)

	// Delete removes a service binding from an app.
	Delete(serviceInstanceName, appName string, opts ...DeleteOption) error

	// List queries Kubernetes for service bindings.
	List(opts ...ListOption) ([]servicecatalogv1beta1.ServiceBinding, error)
}

// NewClient creates a new client capable of interacting with service catalog
// services.
func NewClient(appsClient apps.Client, svcatClient servicecatalogclient.Interface) ClientInterface {
	return &Client{
		appsClient:  appsClient,
		svcatClient: svcatClient,
	}
}

type Client struct {
	appsClient  apps.Client
	svcatClient servicecatalogclient.Interface
}

// Create binds a service instance to an app.
func (c *Client) Create(serviceInstanceName, appName string, opts ...CreateOption) (*v1alpha1.AppSpecServiceBinding, error) {
	cfg := CreateOptionDefaults().Extend(opts).toConfig()

	if serviceInstanceName == "" {
		return nil, errors.New("can't create service binding, no service instance given")
	}

	if appName == "" {
		return nil, errors.New("can't create service binding, no app name given")
	}

	bindingName := cfg.BindingName

	parameters, err := json.Marshal(cfg.Params)
	if err != nil {
		return nil, err
	}

	binding := &v1alpha1.AppSpecServiceBinding{
		Instance:    serviceInstanceName,
		Parameters:  parameters,
		BindingName: bindingName,
	}
	err = c.appsClient.Transform(cfg.Namespace, appName, func(app *v1alpha1.App) error {
		BindService(app, binding)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return binding, nil
}

// Delete unbinds a service instance from an app.
func (c *Client) Delete(serviceInstanceName, appName string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()
	return c.appsClient.Transform(cfg.Namespace, appName, func(app *v1alpha1.App) error {
		UnbindService(app, serviceInstanceName)
		return nil
	})
}

// List queries Kubernetes for service bindings.
func (c *Client) List(opts ...ListOption) ([]servicecatalogv1beta1.ServiceBinding, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	bindings, err := c.svcatClient.
		ServicecatalogV1beta1().
		ServiceBindings(cfg.Namespace).
		List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Filter the results
	filterByServiceInstance := cfg.ServiceInstance != ""
	filterByAppName := cfg.AppName != ""

	var filtered []servicecatalogv1beta1.ServiceBinding
	for _, binding := range bindings.Items {
		if filterByServiceInstance && binding.Spec.InstanceRef.Name != cfg.ServiceInstance {
			continue
		}

		// NOTE: this _could_ be done with a label selector, but we'll do it here
		// to reduce the cognitive overhead of filtering in multiple locations.
		if filterByAppName && binding.Labels[v1alpha1.NameLabel] != cfg.AppName {
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

// BindService binds a service to an App.
func BindService(app *v1alpha1.App, binding *v1alpha1.AppSpecServiceBinding) {
	for i, b := range app.Spec.ServiceBindings {
		if b.BindingName == binding.BindingName {
			app.Spec.ServiceBindings[i] = *binding
			return
		}
	}
	app.Spec.ServiceBindings = append(app.Spec.ServiceBindings, *binding)
}

// UnbindService unbinds a service from an App.
func UnbindService(app *v1alpha1.App, bindingName string) {
	for i, binding := range app.Spec.ServiceBindings {
		if binding.BindingName == bindingName {
			app.Spec.ServiceBindings = append(app.Spec.ServiceBindings[:i], app.Spec.ServiceBindings[i+1:]...)
			break
		}
	}
}
