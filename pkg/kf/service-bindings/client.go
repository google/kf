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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml options.go

// ClientInterface is a client capable of interacting with service catalog services
// and mapping the CF to Kubernetes concepts.
type ClientInterface interface {
	// List queries Kubernetes for service bindings.
	List(opts ...ListOption) ([]servicecatalogv1beta1.ServiceBinding, error)
}

// NewClient creates a new client capable of interacting with service catalog
// services.
func NewClient(svcatClient servicecatalogclient.Interface) ClientInterface {
	return &Client{
		svcatClient: svcatClient,
	}
}

type Client struct {
	svcatClient servicecatalogclient.Interface
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
