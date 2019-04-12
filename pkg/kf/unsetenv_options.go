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

type unsetEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// UnsetEnvOption is a single option for configuring a unsetEnvConfig
type UnsetEnvOption func(*unsetEnvConfig)

// UnsetEnvOptions is a configuration set defining a unsetEnvConfig
type UnsetEnvOptions []UnsetEnvOption

// toConfig applies all the options to a new unsetEnvConfig and returns it.
func (opts UnsetEnvOptions) toConfig() unsetEnvConfig {
	cfg := unsetEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new UnsetEnvOptions with the contents of other overriding
// the values set in this UnsetEnvOptions.
func (opts UnsetEnvOptions) Extend(other UnsetEnvOptions) UnsetEnvOptions {
	var out UnsetEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts UnsetEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithUnsetEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithUnsetEnvNamespace(val string) UnsetEnvOption {
	return func(cfg *unsetEnvConfig) {
		cfg.Namespace = val
	}
}

// UnsetEnvOptionDefaults gets the default values for UnsetEnv.
func UnsetEnvOptionDefaults() UnsetEnvOptions {
	return UnsetEnvOptions{
		WithUnsetEnvNamespace("default"),
	}
}
