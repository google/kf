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
