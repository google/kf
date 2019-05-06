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

package builds

type statusConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// StatusOption is a single option for configuring a statusConfig
type StatusOption func(*statusConfig)

// StatusOptions is a configuration set defining a statusConfig
type StatusOptions []StatusOption

// toConfig applies all the options to a new statusConfig and returns it.
func (opts StatusOptions) toConfig() statusConfig {
	cfg := statusConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new StatusOptions with the contents of other overriding
// the values set in this StatusOptions.
func (opts StatusOptions) Extend(other StatusOptions) StatusOptions {
	var out StatusOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts StatusOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithStatusNamespace creates an Option that sets the Kubernetes namespace to use
func WithStatusNamespace(val string) StatusOption {
	return func(cfg *statusConfig) {
		cfg.Namespace = val
	}
}

// StatusOptionDefaults gets the default values for Status.
func StatusOptionDefaults() StatusOptions {
	return StatusOptions{
		WithStatusNamespace("default"),
	}
}
