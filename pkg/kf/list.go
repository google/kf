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

package kf

import (
	"fmt"

	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// lister lists deployed applications. It should be created via NewLister.
type lister struct {
	c cserving.ServingV1alpha1Interface
}

// NewLister creates a new AppLister.
func NewLister(c cserving.ServingV1alpha1Interface) AppLister {
	return &lister{
		c: c,
	}
}

// List lists the deployed applications for the given namespace.
func (l *lister) List(opts ...ListOption) ([]serving.Service, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()
	listOptions, err := l.setup(cfg, opts)
	if err != nil {
		return nil, err
	}

	services, err := l.c.Services(cfg.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	services.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "knative.dev",
		Version: "v1alpha1",
		Kind:    "Service",
	})

	return services.Items, nil
}

// ListConfigurations lists the deployed Configurations for a given namespace.
func (l *lister) ListConfigurations(opts ...ListConfigurationsOption) ([]serving.Configuration, error) {
	cfg := ListConfigurationsOptionDefaults().Extend(opts).toConfig()
	listOptions, err := l.setupConfiguration(cfg, opts)
	if err != nil {
		return nil, err
	}

	configs, err := l.c.Configurations(cfg.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	return configs.Items, nil
}

// ExtractOneService is a utility function to wrap AppLister.List. It expects
// the results to be exactly one serving.Service, otherwise it will return
// an error.
func ExtractOneService(services []serving.Service, err error) (*serving.Service, error) {
	if err != nil {
		return nil, err
	}

	if len(services) != 1 {
		return nil, fmt.Errorf("expected 1 app, but found %d", len(services))
	}

	return &services[0], nil
}

func (l *lister) setup(cfg listConfig, opts []ListOption) (v1.ListOptions, error) {
	listOptions := v1.ListOptions{}
	if cfg.AppName != "" {
		listOptions.FieldSelector = "metadata.name=" + cfg.AppName
	}

	return listOptions, nil
}

func (l *lister) setupConfiguration(cfg listConfigurationsConfig, opts []ListConfigurationsOption) (v1.ListOptions, error) {
	listOptions := v1.ListOptions{}
	if cfg.AppName != "" {
		listOptions.FieldSelector = "metadata.name=" + cfg.AppName
	}

	return listOptions, nil
}
