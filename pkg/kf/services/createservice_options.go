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
