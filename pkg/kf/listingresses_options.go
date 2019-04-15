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

// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

type listIngressesConfig struct {
	// Namespace is the Kubernetes namespace to search for Ingresses
	Namespace string
	// Service is the name of the ingress service
	Service string
}

// ListIngressesOption is a single option for configuring a listIngressesConfig
type ListIngressesOption func(*listIngressesConfig)

// ListIngressesOptions is a configuration set defining a listIngressesConfig
type ListIngressesOptions []ListIngressesOption

// toConfig applies all the options to a new listIngressesConfig and returns it.
func (opts ListIngressesOptions) toConfig() listIngressesConfig {
	cfg := listIngressesConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListIngressesOptions with the contents of other overriding
// the values set in this ListIngressesOptions.
func (opts ListIngressesOptions) Extend(other ListIngressesOptions) ListIngressesOptions {
	var out ListIngressesOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListIngressesOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Service returns the last set value for Service or the empty value
// if not set.
func (opts ListIngressesOptions) Service() string {
	return opts.toConfig().Service
}

// WithListIngressesNamespace creates an Option that sets the Kubernetes namespace to search for Ingresses
func WithListIngressesNamespace(val string) ListIngressesOption {
	return func(cfg *listIngressesConfig) {
		cfg.Namespace = val
	}
}

// WithListIngressesService creates an Option that sets the name of the ingress service
func WithListIngressesService(val string) ListIngressesOption {
	return func(cfg *listIngressesConfig) {
		cfg.Service = val
	}
}

// ListIngressesOptionDefaults gets the default values for ListIngresses.
func ListIngressesOptionDefaults() ListIngressesOptions {
	return ListIngressesOptions{
		WithListIngressesNamespace("istio-system"),
		WithListIngressesService("istio-ingressgateway"),
	}
}
