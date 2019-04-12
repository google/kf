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

type listConfigurationsConfig struct {
	// AppName is the specific app name to look for
	AppName string
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// ListConfigurationsOption is a single option for configuring a listConfigurationsConfig
type ListConfigurationsOption func(*listConfigurationsConfig)

// ListConfigurationsOptions is a configuration set defining a listConfigurationsConfig
type ListConfigurationsOptions []ListConfigurationsOption

// toConfig applies all the options to a new listConfigurationsConfig and returns it.
func (opts ListConfigurationsOptions) toConfig() listConfigurationsConfig {
	cfg := listConfigurationsConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListConfigurationsOptions with the contents of other overriding
// the values set in this ListConfigurationsOptions.
func (opts ListConfigurationsOptions) Extend(other ListConfigurationsOptions) ListConfigurationsOptions {
	var out ListConfigurationsOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// AppName returns the last set value for AppName or the empty value
// if not set.
func (opts ListConfigurationsOptions) AppName() string {
	return opts.toConfig().AppName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListConfigurationsOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListConfigurationsAppName creates an Option that sets the specific app name to look for
func WithListConfigurationsAppName(val string) ListConfigurationsOption {
	return func(cfg *listConfigurationsConfig) {
		cfg.AppName = val
	}
}

// WithListConfigurationsNamespace creates an Option that sets the Kubernetes namespace to use
func WithListConfigurationsNamespace(val string) ListConfigurationsOption {
	return func(cfg *listConfigurationsConfig) {
		cfg.Namespace = val
	}
}

// ListConfigurationsOptionDefaults gets the default values for ListConfigurations.
func ListConfigurationsOptionDefaults() ListConfigurationsOptions {
	return ListConfigurationsOptions{
		WithListConfigurationsNamespace("default"),
	}
}
