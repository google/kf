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
