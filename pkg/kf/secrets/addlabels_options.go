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

package secrets

type addLabelsConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// AddLabelsOption is a single option for configuring a addLabelsConfig
type AddLabelsOption func(*addLabelsConfig)

// AddLabelsOptions is a configuration set defining a addLabelsConfig
type AddLabelsOptions []AddLabelsOption

// toConfig applies all the options to a new addLabelsConfig and returns it.
func (opts AddLabelsOptions) toConfig() addLabelsConfig {
	cfg := addLabelsConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new AddLabelsOptions with the contents of other overriding
// the values set in this AddLabelsOptions.
func (opts AddLabelsOptions) Extend(other AddLabelsOptions) AddLabelsOptions {
	var out AddLabelsOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts AddLabelsOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithAddLabelsNamespace creates an Option that sets the Kubernetes namespace to use
func WithAddLabelsNamespace(val string) AddLabelsOption {
	return func(cfg *addLabelsConfig) {
		cfg.Namespace = val
	}
}

// AddLabelsOptionDefaults gets the default values for AddLabels.
func AddLabelsOptionDefaults() AddLabelsOptions {
	return AddLabelsOptions{
		WithAddLabelsNamespace("default"),
	}
}
