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

type createConfig struct {
	// Data is data to store in the secret. Values MUST be base64.
	Data map[string][]byte
	// Labels is labels to set on the secret.
	Labels map[string]string
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// StringData is data to store in the secret. Values are encoded in base64 automatically.
	StringData map[string]string
}

// CreateOption is a single option for configuring a createConfig
type CreateOption func(*createConfig)

// CreateOptions is a configuration set defining a createConfig
type CreateOptions []CreateOption

// toConfig applies all the options to a new createConfig and returns it.
func (opts CreateOptions) toConfig() createConfig {
	cfg := createConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new CreateOptions with the contents of other overriding
// the values set in this CreateOptions.
func (opts CreateOptions) Extend(other CreateOptions) CreateOptions {
	var out CreateOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Data returns the last set value for Data or the empty value
// if not set.
func (opts CreateOptions) Data() map[string][]byte {
	return opts.toConfig().Data
}

// Labels returns the last set value for Labels or the empty value
// if not set.
func (opts CreateOptions) Labels() map[string]string {
	return opts.toConfig().Labels
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// StringData returns the last set value for StringData or the empty value
// if not set.
func (opts CreateOptions) StringData() map[string]string {
	return opts.toConfig().StringData
}

// WithCreateData creates an Option that sets data to store in the secret. Values MUST be base64.
func WithCreateData(val map[string][]byte) CreateOption {
	return func(cfg *createConfig) {
		cfg.Data = val
	}
}

// WithCreateLabels creates an Option that sets labels to set on the secret.
func WithCreateLabels(val map[string]string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Labels = val
	}
}

// WithCreateNamespace creates an Option that sets the Kubernetes namespace to use
func WithCreateNamespace(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Namespace = val
	}
}

// WithCreateStringData creates an Option that sets data to store in the secret. Values are encoded in base64 automatically.
func WithCreateStringData(val map[string]string) CreateOption {
	return func(cfg *createConfig) {
		cfg.StringData = val
	}
}

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{
		WithCreateNamespace("default"),
	}
}

type deleteConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// DeleteOption is a single option for configuring a deleteConfig
type DeleteOption func(*deleteConfig)

// DeleteOptions is a configuration set defining a deleteConfig
type DeleteOptions []DeleteOption

// toConfig applies all the options to a new deleteConfig and returns it.
func (opts DeleteOptions) toConfig() deleteConfig {
	cfg := deleteConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteOptions with the contents of other overriding
// the values set in this DeleteOptions.
func (opts DeleteOptions) Extend(other DeleteOptions) DeleteOptions {
	var out DeleteOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeleteOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeleteNamespace creates an Option that sets the Kubernetes namespace to use
func WithDeleteNamespace(val string) DeleteOption {
	return func(cfg *deleteConfig) {
		cfg.Namespace = val
	}
}

// DeleteOptionDefaults gets the default values for Delete.
func DeleteOptionDefaults() DeleteOptions {
	return DeleteOptions{
		WithDeleteNamespace("default"),
	}
}

type getConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// GetOption is a single option for configuring a getConfig
type GetOption func(*getConfig)

// GetOptions is a configuration set defining a getConfig
type GetOptions []GetOption

// toConfig applies all the options to a new getConfig and returns it.
func (opts GetOptions) toConfig() getConfig {
	cfg := getConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new GetOptions with the contents of other overriding
// the values set in this GetOptions.
func (opts GetOptions) Extend(other GetOptions) GetOptions {
	var out GetOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts GetOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithGetNamespace creates an Option that sets the Kubernetes namespace to use
func WithGetNamespace(val string) GetOption {
	return func(cfg *getConfig) {
		cfg.Namespace = val
	}
}

// GetOptionDefaults gets the default values for Get.
func GetOptionDefaults() GetOptions {
	return GetOptions{
		WithGetNamespace("default"),
	}
}

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

type listConfig struct {
	// LabelSelector is filters results to only labels matching the filter.
	LabelSelector string
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// ListOption is a single option for configuring a listConfig
type ListOption func(*listConfig)

// ListOptions is a configuration set defining a listConfig
type ListOptions []ListOption

// toConfig applies all the options to a new listConfig and returns it.
func (opts ListOptions) toConfig() listConfig {
	cfg := listConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListOptions with the contents of other overriding
// the values set in this ListOptions.
func (opts ListOptions) Extend(other ListOptions) ListOptions {
	var out ListOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// LabelSelector returns the last set value for LabelSelector or the empty value
// if not set.
func (opts ListOptions) LabelSelector() string {
	return opts.toConfig().LabelSelector
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListLabelSelector creates an Option that sets filters results to only labels matching the filter.
func WithListLabelSelector(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.LabelSelector = val
	}
}

// WithListNamespace creates an Option that sets the Kubernetes namespace to use
func WithListNamespace(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Namespace = val
	}
}

// ListOptionDefaults gets the default values for List.
func ListOptionDefaults() ListOptions {
	return ListOptions{
		WithListNamespace("default"),
	}
}
