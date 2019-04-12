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

type setEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// SetEnvOption is a single option for configuring a setEnvConfig
type SetEnvOption func(*setEnvConfig)

// SetEnvOptions is a configuration set defining a setEnvConfig
type SetEnvOptions []SetEnvOption

// toConfig applies all the options to a new setEnvConfig and returns it.
func (opts SetEnvOptions) toConfig() setEnvConfig {
	cfg := setEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new SetEnvOptions with the contents of other overriding
// the values set in this SetEnvOptions.
func (opts SetEnvOptions) Extend(other SetEnvOptions) SetEnvOptions {
	var out SetEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts SetEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithSetEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithSetEnvNamespace(val string) SetEnvOption {
	return func(cfg *setEnvConfig) {
		cfg.Namespace = val
	}
}

// SetEnvOptionDefaults gets the default values for SetEnv.
func SetEnvOptionDefaults() SetEnvOptions {
	return SetEnvOptions{
		WithSetEnvNamespace("default"),
	}
}
