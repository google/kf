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

package buildpacks

type uploadBuildTemplateConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// UploadBuildTemplateOption is a single option for configuring a uploadBuildTemplateConfig
type UploadBuildTemplateOption func(*uploadBuildTemplateConfig)

// UploadBuildTemplateOptions is a configuration set defining a uploadBuildTemplateConfig
type UploadBuildTemplateOptions []UploadBuildTemplateOption

// toConfig applies all the options to a new uploadBuildTemplateConfig and returns it.
func (opts UploadBuildTemplateOptions) toConfig() uploadBuildTemplateConfig {
	cfg := uploadBuildTemplateConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new UploadBuildTemplateOptions with the contents of other overriding
// the values set in this UploadBuildTemplateOptions.
func (opts UploadBuildTemplateOptions) Extend(other UploadBuildTemplateOptions) UploadBuildTemplateOptions {
	var out UploadBuildTemplateOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts UploadBuildTemplateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithUploadBuildTemplateNamespace creates an Option that sets the Kubernetes namespace to use
func WithUploadBuildTemplateNamespace(val string) UploadBuildTemplateOption {
	return func(cfg *uploadBuildTemplateConfig) {
		cfg.Namespace = val
	}
}

// UploadBuildTemplateOptionDefaults gets the default values for UploadBuildTemplate.
func UploadBuildTemplateOptionDefaults() UploadBuildTemplateOptions {
	return UploadBuildTemplateOptions{
		WithUploadBuildTemplateNamespace("default"),
	}
}
