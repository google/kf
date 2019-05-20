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

package services

type createServiceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// Params is service-specific configuration parameters.
	Params map[string]interface{}
}

// CreateServiceOption is a single option for configuring a createServiceConfig
type CreateServiceOption func(*createServiceConfig)

// CreateServiceOptions is a configuration set defining a createServiceConfig
type CreateServiceOptions []CreateServiceOption

// toConfig applies all the options to a new createServiceConfig and returns it.
func (opts CreateServiceOptions) toConfig() createServiceConfig {
	cfg := createServiceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new CreateServiceOptions with the contents of other overriding
// the values set in this CreateServiceOptions.
func (opts CreateServiceOptions) Extend(other CreateServiceOptions) CreateServiceOptions {
	var out CreateServiceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateServiceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Params returns the last set value for Params or the empty value
// if not set.
func (opts CreateServiceOptions) Params() map[string]interface{} {
	return opts.toConfig().Params
}

// WithCreateServiceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithCreateServiceNamespace(val string) CreateServiceOption {
	return func(cfg *createServiceConfig) {
		cfg.Namespace = val
	}
}

// WithCreateServiceParams creates an Option that sets service-specific configuration parameters.
func WithCreateServiceParams(val map[string]interface{}) CreateServiceOption {
	return func(cfg *createServiceConfig) {
		cfg.Params = val
	}
}

// CreateServiceOptionDefaults gets the default values for CreateService.
func CreateServiceOptionDefaults() CreateServiceOptions {
	return CreateServiceOptions{
		WithCreateServiceNamespace("default"),
	}
}

type deleteServiceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// DeleteServiceOption is a single option for configuring a deleteServiceConfig
type DeleteServiceOption func(*deleteServiceConfig)

// DeleteServiceOptions is a configuration set defining a deleteServiceConfig
type DeleteServiceOptions []DeleteServiceOption

// toConfig applies all the options to a new deleteServiceConfig and returns it.
func (opts DeleteServiceOptions) toConfig() deleteServiceConfig {
	cfg := deleteServiceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteServiceOptions with the contents of other overriding
// the values set in this DeleteServiceOptions.
func (opts DeleteServiceOptions) Extend(other DeleteServiceOptions) DeleteServiceOptions {
	var out DeleteServiceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeleteServiceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeleteServiceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithDeleteServiceNamespace(val string) DeleteServiceOption {
	return func(cfg *deleteServiceConfig) {
		cfg.Namespace = val
	}
}

// DeleteServiceOptionDefaults gets the default values for DeleteService.
func DeleteServiceOptionDefaults() DeleteServiceOptions {
	return DeleteServiceOptions{
		WithDeleteServiceNamespace("default"),
	}
}

type getServiceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// GetServiceOption is a single option for configuring a getServiceConfig
type GetServiceOption func(*getServiceConfig)

// GetServiceOptions is a configuration set defining a getServiceConfig
type GetServiceOptions []GetServiceOption

// toConfig applies all the options to a new getServiceConfig and returns it.
func (opts GetServiceOptions) toConfig() getServiceConfig {
	cfg := getServiceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new GetServiceOptions with the contents of other overriding
// the values set in this GetServiceOptions.
func (opts GetServiceOptions) Extend(other GetServiceOptions) GetServiceOptions {
	var out GetServiceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts GetServiceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithGetServiceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithGetServiceNamespace(val string) GetServiceOption {
	return func(cfg *getServiceConfig) {
		cfg.Namespace = val
	}
}

// GetServiceOptionDefaults gets the default values for GetService.
func GetServiceOptionDefaults() GetServiceOptions {
	return GetServiceOptions{
		WithGetServiceNamespace("default"),
	}
}

type listServicesConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// ListServicesOption is a single option for configuring a listServicesConfig
type ListServicesOption func(*listServicesConfig)

// ListServicesOptions is a configuration set defining a listServicesConfig
type ListServicesOptions []ListServicesOption

// toConfig applies all the options to a new listServicesConfig and returns it.
func (opts ListServicesOptions) toConfig() listServicesConfig {
	cfg := listServicesConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListServicesOptions with the contents of other overriding
// the values set in this ListServicesOptions.
func (opts ListServicesOptions) Extend(other ListServicesOptions) ListServicesOptions {
	var out ListServicesOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListServicesOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListServicesNamespace creates an Option that sets the Kubernetes namespace to use.
func WithListServicesNamespace(val string) ListServicesOption {
	return func(cfg *listServicesConfig) {
		cfg.Namespace = val
	}
}

// ListServicesOptionDefaults gets the default values for ListServices.
func ListServicesOptionDefaults() ListServicesOptions {
	return ListServicesOptions{
		WithListServicesNamespace("default"),
	}
}

type marketplaceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// MarketplaceOption is a single option for configuring a marketplaceConfig
type MarketplaceOption func(*marketplaceConfig)

// MarketplaceOptions is a configuration set defining a marketplaceConfig
type MarketplaceOptions []MarketplaceOption

// toConfig applies all the options to a new marketplaceConfig and returns it.
func (opts MarketplaceOptions) toConfig() marketplaceConfig {
	cfg := marketplaceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new MarketplaceOptions with the contents of other overriding
// the values set in this MarketplaceOptions.
func (opts MarketplaceOptions) Extend(other MarketplaceOptions) MarketplaceOptions {
	var out MarketplaceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts MarketplaceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithMarketplaceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithMarketplaceNamespace(val string) MarketplaceOption {
	return func(cfg *marketplaceConfig) {
		cfg.Namespace = val
	}
}

// MarketplaceOptionDefaults gets the default values for Marketplace.
func MarketplaceOptionDefaults() MarketplaceOptions {
	return MarketplaceOptions{
		WithMarketplaceNamespace("default"),
	}
}
