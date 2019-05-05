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

package generator

type registerConfig struct {
	// UploadDir is upload a directory as part of the command
	UploadDir bool
}

// RegisterOption is a single option for configuring a registerConfig
type RegisterOption func(*registerConfig)

// RegisterOptions is a configuration set defining a registerConfig
type RegisterOptions []RegisterOption

// toConfig applies all the options to a new registerConfig and returns it.
func (opts RegisterOptions) toConfig() registerConfig {
	cfg := registerConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new RegisterOptions with the contents of other overriding
// the values set in this RegisterOptions.
func (opts RegisterOptions) Extend(other RegisterOptions) RegisterOptions {
	var out RegisterOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// UploadDir returns the last set value for UploadDir or the empty value
// if not set.
func (opts RegisterOptions) UploadDir() bool {
	return opts.toConfig().UploadDir
}

// WithRegisterUploadDir creates an Option that sets upload a directory as part of the command
func WithRegisterUploadDir(val bool) RegisterOption {
	return func(cfg *registerConfig) {
		cfg.UploadDir = val
	}
}

// RegisterOptionDefaults gets the default values for Register.
func RegisterOptionDefaults() RegisterOptions {
	return RegisterOptions{}
}
