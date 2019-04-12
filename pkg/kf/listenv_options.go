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

type listEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// ListEnvOption is a single option for configuring a listEnvConfig
type ListEnvOption func(*listEnvConfig)

// ListEnvOptions is a configuration set defining a listEnvConfig
type ListEnvOptions []ListEnvOption

// toConfig applies all the options to a new listEnvConfig and returns it.
func (opts ListEnvOptions) toConfig() listEnvConfig {
	cfg := listEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListEnvOptions with the contents of other overriding
// the values set in this ListEnvOptions.
func (opts ListEnvOptions) Extend(other ListEnvOptions) ListEnvOptions {
	var out ListEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithListEnvNamespace(val string) ListEnvOption {
	return func(cfg *listEnvConfig) {
		cfg.Namespace = val
	}
}

// ListEnvOptionDefaults gets the default values for ListEnv.
func ListEnvOptionDefaults() ListEnvOptions {
	return ListEnvOptions{
		WithListEnvNamespace("default"),
	}
}
