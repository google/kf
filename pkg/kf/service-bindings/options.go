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

package servicebindings

type createConfig struct {
	// BindingName is name to expose service instance to app process with.
	BindingName string
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// Params is service-specific configuration parameters.
	Params map[string]interface{}
}

// CreateOption is a single option for configuring a createConfig
type CreateOption func(*createConfig)

// CreateOptions is a configuration set defining a createConfig
type CreateOptions []CreateOption

// toConfig applies all the options to a new createConfig and returns it.
func (opts CreateOptions) toConfig() createConfig {
	cfg := createConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new CreateOptions with the contents of other overriding
// the values set in this CreateOptions.
func (opts CreateOptions) Extend(other CreateOptions) CreateOptions {
	var out CreateOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// BindingName returns the last set value for BindingName or the empty value
// if not set.
func (opts CreateOptions) BindingName() string {
	return opts.toConfig().BindingName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Params returns the last set value for Params or the empty value
// if not set.
func (opts CreateOptions) Params() map[string]interface{} {
	return opts.toConfig().Params
}

// WithCreateBindingName creates an Option that sets name to expose service instance to app process with.
func WithCreateBindingName(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.BindingName = val
	}
}

// WithCreateNamespace creates an Option that sets the Kubernetes namespace to use.
func WithCreateNamespace(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Namespace = val
	}
}

// WithCreateParams creates an Option that sets service-specific configuration parameters.
func WithCreateParams(val map[string]interface{}) CreateOption {
	return func(cfg *createConfig) {
		cfg.Params = val
	}
}

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{
		WithCreateNamespace("default"),
	}
}

type deleteConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// DeleteOption is a single option for configuring a deleteConfig
type DeleteOption func(*deleteConfig)

// DeleteOptions is a configuration set defining a deleteConfig
type DeleteOptions []DeleteOption

// toConfig applies all the options to a new deleteConfig and returns it.
func (opts DeleteOptions) toConfig() deleteConfig {
	cfg := deleteConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteOptions with the contents of other overriding
// the values set in this DeleteOptions.
func (opts DeleteOptions) Extend(other DeleteOptions) DeleteOptions {
	var out DeleteOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeleteOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeleteNamespace creates an Option that sets the Kubernetes namespace to use.
func WithDeleteNamespace(val string) DeleteOption {
	return func(cfg *deleteConfig) {
		cfg.Namespace = val
	}
}

// DeleteOptionDefaults gets the default values for Delete.
func DeleteOptionDefaults() DeleteOptions {
	return DeleteOptions{
		WithDeleteNamespace("default"),
	}
}

type listConfig struct {
	// AppName is filter the results to bindings for the given app.
	AppName string
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// ServiceInstance is filter the results to bindings for the given service instance.
	ServiceInstance string
}

// ListOption is a single option for configuring a listConfig
type ListOption func(*listConfig)

// ListOptions is a configuration set defining a listConfig
type ListOptions []ListOption

// toConfig applies all the options to a new listConfig and returns it.
func (opts ListOptions) toConfig() listConfig {
	cfg := listConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListOptions with the contents of other overriding
// the values set in this ListOptions.
func (opts ListOptions) Extend(other ListOptions) ListOptions {
	var out ListOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// AppName returns the last set value for AppName or the empty value
// if not set.
func (opts ListOptions) AppName() string {
	return opts.toConfig().AppName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// ServiceInstance returns the last set value for ServiceInstance or the empty value
// if not set.
func (opts ListOptions) ServiceInstance() string {
	return opts.toConfig().ServiceInstance
}

// WithListAppName creates an Option that sets filter the results to bindings for the given app.
func WithListAppName(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.AppName = val
	}
}

// WithListNamespace creates an Option that sets the Kubernetes namespace to use.
func WithListNamespace(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Namespace = val
	}
}

// WithListServiceInstance creates an Option that sets filter the results to bindings for the given service instance.
func WithListServiceInstance(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.ServiceInstance = val
	}
}

// ListOptionDefaults gets the default values for List.
func ListOptionDefaults() ListOptions {
	return ListOptions{
		WithListNamespace("default"),
	}
}

type getVcapServicesConfig struct {
	// FailOnBadSecret is fail if a binding refers to an invalid (or not yet created) secret.
	FailOnBadSecret bool
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// GetVcapServicesOption is a single option for configuring a getVcapServicesConfig
type GetVcapServicesOption func(*getVcapServicesConfig)

// GetVcapServicesOptions is a configuration set defining a getVcapServicesConfig
type GetVcapServicesOptions []GetVcapServicesOption

// toConfig applies all the options to a new getVcapServicesConfig and returns it.
func (opts GetVcapServicesOptions) toConfig() getVcapServicesConfig {
	cfg := getVcapServicesConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new GetVcapServicesOptions with the contents of other overriding
// the values set in this GetVcapServicesOptions.
func (opts GetVcapServicesOptions) Extend(other GetVcapServicesOptions) GetVcapServicesOptions {
	var out GetVcapServicesOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// FailOnBadSecret returns the last set value for FailOnBadSecret or the empty value
// if not set.
func (opts GetVcapServicesOptions) FailOnBadSecret() bool {
	return opts.toConfig().FailOnBadSecret
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts GetVcapServicesOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithGetVcapServicesFailOnBadSecret creates an Option that sets fail if a binding refers to an invalid (or not yet created) secret.
func WithGetVcapServicesFailOnBadSecret(val bool) GetVcapServicesOption {
	return func(cfg *getVcapServicesConfig) {
		cfg.FailOnBadSecret = val
	}
}

// WithGetVcapServicesNamespace creates an Option that sets the Kubernetes namespace to use.
func WithGetVcapServicesNamespace(val string) GetVcapServicesOption {
	return func(cfg *getVcapServicesConfig) {
		cfg.Namespace = val
	}
}

// GetVcapServicesOptionDefaults gets the default values for GetVcapServices.
func GetVcapServicesOptionDefaults() GetVcapServicesOptions {
	return GetVcapServicesOptions{
		WithGetVcapServicesFailOnBadSecret(false),
		WithGetVcapServicesNamespace("default"),
	}
}
