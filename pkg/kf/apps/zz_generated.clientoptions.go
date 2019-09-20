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

package apps

type createConfig struct {
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

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{}
}

type updateConfig struct {
}

// UpdateOption is a single option for configuring a updateConfig
type UpdateOption func(*updateConfig)

// UpdateOptions is a configuration set defining a updateConfig
type UpdateOptions []UpdateOption

// toConfig applies all the options to a new updateConfig and returns it.
func (opts UpdateOptions) toConfig() updateConfig {
	cfg := updateConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new UpdateOptions with the contents of other overriding
// the values set in this UpdateOptions.
func (opts UpdateOptions) Extend(other UpdateOptions) UpdateOptions {
	var out UpdateOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// UpdateOptionDefaults gets the default values for Update.
func UpdateOptionDefaults() UpdateOptions {
	return UpdateOptions{}
}

type getConfig struct {
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

// GetOptionDefaults gets the default values for Get.
func GetOptionDefaults() GetOptions {
	return GetOptions{}
}

type deleteConfig struct {
	// ForegroundDeletion is If the resource should be deleted in the foreground.
	ForegroundDeletion bool
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

// ForegroundDeletion returns the last set value for ForegroundDeletion or the empty value
// if not set.
func (opts DeleteOptions) ForegroundDeletion() bool {
	return opts.toConfig().ForegroundDeletion
}

// WithDeleteForegroundDeletion creates an Option that sets If the resource should be deleted in the foreground.
func WithDeleteForegroundDeletion(val bool) DeleteOption {
	return func(cfg *deleteConfig) {
		cfg.ForegroundDeletion = val
	}
}

// DeleteOptionDefaults gets the default values for Delete.
func DeleteOptionDefaults() DeleteOptions {
	return DeleteOptions{}
}

type listConfig struct {
	// fieldSelector is A selector on the resource's fields.
	fieldSelector map[string]string
	// filter is Filter to apply.
	filter Predicate
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

// fieldSelector returns the last set value for fieldSelector or the empty value
// if not set.
func (opts ListOptions) fieldSelector() map[string]string {
	return opts.toConfig().fieldSelector
}

// filter returns the last set value for filter or the empty value
// if not set.
func (opts ListOptions) filter() Predicate {
	return opts.toConfig().filter
}

// WithListFieldSelector creates an Option that sets A selector on the resource's fields.
func WithListFieldSelector(val map[string]string) ListOption {
	return func(cfg *listConfig) {
		cfg.fieldSelector = val
	}
}

// WithListFilter creates an Option that sets Filter to apply.
func WithListFilter(val Predicate) ListOption {
	return func(cfg *listConfig) {
		cfg.filter = val
	}
}

// ListOptionDefaults gets the default values for List.
func ListOptionDefaults() ListOptions {
	return ListOptions{}
}
