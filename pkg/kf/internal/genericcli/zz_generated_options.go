// Copyright 2023 Google LLC
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

package genericcli

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
)

type listConfig struct {
	// Aliases is an array of aliases that can be used instead of the command name.
	Aliases []string
	// ArgumentFilters is callbacks that can modify the lister.
	ArgumentFilters []ListArgumentFilter
	// CommandName is the name to use for the command.
	CommandName string
	// Example is the example to use for the command.
	Example string
	// LabelFilters is flag name to label pairs to use as list filters.
	LabelFilters map[string]string
	// LabelRequirements is label requirements to filter resources.
	LabelRequirements []labels.Requirement
	// Long is the long description to use for the command.
	Long string
	// PluralFriendlyName is the plural object name to display for this resource.
	PluralFriendlyName string
	// Short is the short description to use for the command.
	Short string
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

// Aliases returns the last set value for Aliases or the empty value
// if not set.
func (opts ListOptions) Aliases() []string {
	return opts.toConfig().Aliases
}

// ArgumentFilters returns the last set value for ArgumentFilters or the empty value
// if not set.
func (opts ListOptions) ArgumentFilters() []ListArgumentFilter {
	return opts.toConfig().ArgumentFilters
}

// CommandName returns the last set value for CommandName or the empty value
// if not set.
func (opts ListOptions) CommandName() string {
	return opts.toConfig().CommandName
}

// Example returns the last set value for Example or the empty value
// if not set.
func (opts ListOptions) Example() string {
	return opts.toConfig().Example
}

// LabelFilters returns the last set value for LabelFilters or the empty value
// if not set.
func (opts ListOptions) LabelFilters() map[string]string {
	return opts.toConfig().LabelFilters
}

// LabelRequirements returns the last set value for LabelRequirements or the empty value
// if not set.
func (opts ListOptions) LabelRequirements() []labels.Requirement {
	return opts.toConfig().LabelRequirements
}

// Long returns the last set value for Long or the empty value
// if not set.
func (opts ListOptions) Long() string {
	return opts.toConfig().Long
}

// PluralFriendlyName returns the last set value for PluralFriendlyName or the empty value
// if not set.
func (opts ListOptions) PluralFriendlyName() string {
	return opts.toConfig().PluralFriendlyName
}

// Short returns the last set value for Short or the empty value
// if not set.
func (opts ListOptions) Short() string {
	return opts.toConfig().Short
}

// WithListAliases creates an Option that sets an array of aliases that can be used instead of the command name.
func WithListAliases(val []string) ListOption {
	return func(cfg *listConfig) {
		cfg.Aliases = val
	}
}

// WithListArgumentFilters creates an Option that sets callbacks that can modify the lister.
func WithListArgumentFilters(val []ListArgumentFilter) ListOption {
	return func(cfg *listConfig) {
		cfg.ArgumentFilters = val
	}
}

// WithListCommandName creates an Option that sets the name to use for the command.
func WithListCommandName(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.CommandName = val
	}
}

// WithListExample creates an Option that sets the example to use for the command.
func WithListExample(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Example = val
	}
}

// WithListLabelFilters creates an Option that sets flag name to label pairs to use as list filters.
func WithListLabelFilters(val map[string]string) ListOption {
	return func(cfg *listConfig) {
		cfg.LabelFilters = val
	}
}

// WithListLabelRequirements creates an Option that sets label requirements to filter resources.
func WithListLabelRequirements(val []labels.Requirement) ListOption {
	return func(cfg *listConfig) {
		cfg.LabelRequirements = val
	}
}

// WithListLong creates an Option that sets the long description to use for the command.
func WithListLong(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Long = val
	}
}

// WithListPluralFriendlyName creates an Option that sets the plural object name to display for this resource.
func WithListPluralFriendlyName(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.PluralFriendlyName = val
	}
}

// WithListShort creates an Option that sets the short description to use for the command.
func WithListShort(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Short = val
	}
}

// ListOptionDefaults gets the default values for List.
func ListOptionDefaults() ListOptions {
	return ListOptions{}
}

type describeConfig struct {
	// Aliases is an array of aliases that can be used instead of the command name.
	Aliases []string
	// CommandName is the name to use for the command.
	CommandName string
	// Example is the example to use for the command.
	Example string
	// Long is the long description to use for the command.
	Long string
	// Short is the short description to use for the command.
	Short string
}

// DescribeOption is a single option for configuring a describeConfig
type DescribeOption func(*describeConfig)

// DescribeOptions is a configuration set defining a describeConfig
type DescribeOptions []DescribeOption

// toConfig applies all the options to a new describeConfig and returns it.
func (opts DescribeOptions) toConfig() describeConfig {
	cfg := describeConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DescribeOptions with the contents of other overriding
// the values set in this DescribeOptions.
func (opts DescribeOptions) Extend(other DescribeOptions) DescribeOptions {
	var out DescribeOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Aliases returns the last set value for Aliases or the empty value
// if not set.
func (opts DescribeOptions) Aliases() []string {
	return opts.toConfig().Aliases
}

// CommandName returns the last set value for CommandName or the empty value
// if not set.
func (opts DescribeOptions) CommandName() string {
	return opts.toConfig().CommandName
}

// Example returns the last set value for Example or the empty value
// if not set.
func (opts DescribeOptions) Example() string {
	return opts.toConfig().Example
}

// Long returns the last set value for Long or the empty value
// if not set.
func (opts DescribeOptions) Long() string {
	return opts.toConfig().Long
}

// Short returns the last set value for Short or the empty value
// if not set.
func (opts DescribeOptions) Short() string {
	return opts.toConfig().Short
}

// WithDescribeAliases creates an Option that sets an array of aliases that can be used instead of the command name.
func WithDescribeAliases(val []string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.Aliases = val
	}
}

// WithDescribeCommandName creates an Option that sets the name to use for the command.
func WithDescribeCommandName(val string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.CommandName = val
	}
}

// WithDescribeExample creates an Option that sets the example to use for the command.
func WithDescribeExample(val string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.Example = val
	}
}

// WithDescribeLong creates an Option that sets the long description to use for the command.
func WithDescribeLong(val string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.Long = val
	}
}

// WithDescribeShort creates an Option that sets the short description to use for the command.
func WithDescribeShort(val string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.Short = val
	}
}

// DescribeOptionDefaults gets the default values for Describe.
func DescribeOptionDefaults() DescribeOptions {
	return DescribeOptions{}
}

type stubConfig struct {
	// Aliases is an array of aliases that can be used instead of the command name.
	Aliases []string
	// CommandName is the name to use for the command.
	CommandName string
	// Example is the example to use for the command.
	Example string
	// Long is the long description to use for the command.
	Long string
	// Short is the short description to use for the command.
	Short string
}

// StubOption is a single option for configuring a stubConfig
type StubOption func(*stubConfig)

// StubOptions is a configuration set defining a stubConfig
type StubOptions []StubOption

// toConfig applies all the options to a new stubConfig and returns it.
func (opts StubOptions) toConfig() stubConfig {
	cfg := stubConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new StubOptions with the contents of other overriding
// the values set in this StubOptions.
func (opts StubOptions) Extend(other StubOptions) StubOptions {
	var out StubOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Aliases returns the last set value for Aliases or the empty value
// if not set.
func (opts StubOptions) Aliases() []string {
	return opts.toConfig().Aliases
}

// CommandName returns the last set value for CommandName or the empty value
// if not set.
func (opts StubOptions) CommandName() string {
	return opts.toConfig().CommandName
}

// Example returns the last set value for Example or the empty value
// if not set.
func (opts StubOptions) Example() string {
	return opts.toConfig().Example
}

// Long returns the last set value for Long or the empty value
// if not set.
func (opts StubOptions) Long() string {
	return opts.toConfig().Long
}

// Short returns the last set value for Short or the empty value
// if not set.
func (opts StubOptions) Short() string {
	return opts.toConfig().Short
}

// WithStubAliases creates an Option that sets an array of aliases that can be used instead of the command name.
func WithStubAliases(val []string) StubOption {
	return func(cfg *stubConfig) {
		cfg.Aliases = val
	}
}

// WithStubCommandName creates an Option that sets the name to use for the command.
func WithStubCommandName(val string) StubOption {
	return func(cfg *stubConfig) {
		cfg.CommandName = val
	}
}

// WithStubExample creates an Option that sets the example to use for the command.
func WithStubExample(val string) StubOption {
	return func(cfg *stubConfig) {
		cfg.Example = val
	}
}

// WithStubLong creates an Option that sets the long description to use for the command.
func WithStubLong(val string) StubOption {
	return func(cfg *stubConfig) {
		cfg.Long = val
	}
}

// WithStubShort creates an Option that sets the short description to use for the command.
func WithStubShort(val string) StubOption {
	return func(cfg *stubConfig) {
		cfg.Short = val
	}
}

// StubOptionDefaults gets the default values for Stub.
func StubOptionDefaults() StubOptions {
	return StubOptions{}
}

type deleteByNameConfig struct {
	// AdditionalLongText is additional text to append to long.
	AdditionalLongText string
	// Aliases is an array of aliases that can be used instead of the command name.
	Aliases []string
	// CommandName is the name to use for the command.
	CommandName string
	// Example is the example to use for the command.
	Example string
	// Long is the long description to use for the command.
	Long string
	// PropagationPolicy is propagation policy for deleting an object.
	PropagationPolicy metav1.DeletionPropagation
	// Short is the short description to use for the command.
	Short string
}

// DeleteByNameOption is a single option for configuring a deleteByNameConfig
type DeleteByNameOption func(*deleteByNameConfig)

// DeleteByNameOptions is a configuration set defining a deleteByNameConfig
type DeleteByNameOptions []DeleteByNameOption

// toConfig applies all the options to a new deleteByNameConfig and returns it.
func (opts DeleteByNameOptions) toConfig() deleteByNameConfig {
	cfg := deleteByNameConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteByNameOptions with the contents of other overriding
// the values set in this DeleteByNameOptions.
func (opts DeleteByNameOptions) Extend(other DeleteByNameOptions) DeleteByNameOptions {
	var out DeleteByNameOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// AdditionalLongText returns the last set value for AdditionalLongText or the empty value
// if not set.
func (opts DeleteByNameOptions) AdditionalLongText() string {
	return opts.toConfig().AdditionalLongText
}

// Aliases returns the last set value for Aliases or the empty value
// if not set.
func (opts DeleteByNameOptions) Aliases() []string {
	return opts.toConfig().Aliases
}

// CommandName returns the last set value for CommandName or the empty value
// if not set.
func (opts DeleteByNameOptions) CommandName() string {
	return opts.toConfig().CommandName
}

// Example returns the last set value for Example or the empty value
// if not set.
func (opts DeleteByNameOptions) Example() string {
	return opts.toConfig().Example
}

// Long returns the last set value for Long or the empty value
// if not set.
func (opts DeleteByNameOptions) Long() string {
	return opts.toConfig().Long
}

// PropagationPolicy returns the last set value for PropagationPolicy or the empty value
// if not set.
func (opts DeleteByNameOptions) PropagationPolicy() metav1.DeletionPropagation {
	return opts.toConfig().PropagationPolicy
}

// Short returns the last set value for Short or the empty value
// if not set.
func (opts DeleteByNameOptions) Short() string {
	return opts.toConfig().Short
}

// WithDeleteByNameAdditionalLongText creates an Option that sets additional text to append to long.
func WithDeleteByNameAdditionalLongText(val string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.AdditionalLongText = val
	}
}

// WithDeleteByNameAliases creates an Option that sets an array of aliases that can be used instead of the command name.
func WithDeleteByNameAliases(val []string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.Aliases = val
	}
}

// WithDeleteByNameCommandName creates an Option that sets the name to use for the command.
func WithDeleteByNameCommandName(val string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.CommandName = val
	}
}

// WithDeleteByNameExample creates an Option that sets the example to use for the command.
func WithDeleteByNameExample(val string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.Example = val
	}
}

// WithDeleteByNameLong creates an Option that sets the long description to use for the command.
func WithDeleteByNameLong(val string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.Long = val
	}
}

// WithDeleteByNamePropagationPolicy creates an Option that sets propagation policy for deleting an object.
func WithDeleteByNamePropagationPolicy(val metav1.DeletionPropagation) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.PropagationPolicy = val
	}
}

// WithDeleteByNameShort creates an Option that sets the short description to use for the command.
func WithDeleteByNameShort(val string) DeleteByNameOption {
	return func(cfg *deleteByNameConfig) {
		cfg.Short = val
	}
}

// DeleteByNameOptionDefaults gets the default values for DeleteByName.
func DeleteByNameOptionDefaults() DeleteByNameOptions {
	return DeleteByNameOptions{}
}

type xargsConfig struct {
	// Aliases is an array of aliases that can be used instead of the command name.
	Aliases []string
	// CommandName is the name to use for the command.
	CommandName string
	// Example is the example to use for the command.
	Example string
	// LabelFilters is flag name to label pairs to use as list filters.
	LabelFilters map[string]string
	// LabelRequirements is label requirements to filter resources.
	LabelRequirements []labels.Requirement
	// Long is the long description to use for the command.
	Long string
	// PluralFriendlyName is the plural object name to display for this resource.
	PluralFriendlyName string
	// Short is the short description to use for the command.
	Short string
}

// XargsOption is a single option for configuring a xargsConfig
type XargsOption func(*xargsConfig)

// XargsOptions is a configuration set defining a xargsConfig
type XargsOptions []XargsOption

// toConfig applies all the options to a new xargsConfig and returns it.
func (opts XargsOptions) toConfig() xargsConfig {
	cfg := xargsConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new XargsOptions with the contents of other overriding
// the values set in this XargsOptions.
func (opts XargsOptions) Extend(other XargsOptions) XargsOptions {
	var out XargsOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Aliases returns the last set value for Aliases or the empty value
// if not set.
func (opts XargsOptions) Aliases() []string {
	return opts.toConfig().Aliases
}

// CommandName returns the last set value for CommandName or the empty value
// if not set.
func (opts XargsOptions) CommandName() string {
	return opts.toConfig().CommandName
}

// Example returns the last set value for Example or the empty value
// if not set.
func (opts XargsOptions) Example() string {
	return opts.toConfig().Example
}

// LabelFilters returns the last set value for LabelFilters or the empty value
// if not set.
func (opts XargsOptions) LabelFilters() map[string]string {
	return opts.toConfig().LabelFilters
}

// LabelRequirements returns the last set value for LabelRequirements or the empty value
// if not set.
func (opts XargsOptions) LabelRequirements() []labels.Requirement {
	return opts.toConfig().LabelRequirements
}

// Long returns the last set value for Long or the empty value
// if not set.
func (opts XargsOptions) Long() string {
	return opts.toConfig().Long
}

// PluralFriendlyName returns the last set value for PluralFriendlyName or the empty value
// if not set.
func (opts XargsOptions) PluralFriendlyName() string {
	return opts.toConfig().PluralFriendlyName
}

// Short returns the last set value for Short or the empty value
// if not set.
func (opts XargsOptions) Short() string {
	return opts.toConfig().Short
}

// WithXargsAliases creates an Option that sets an array of aliases that can be used instead of the command name.
func WithXargsAliases(val []string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.Aliases = val
	}
}

// WithXargsCommandName creates an Option that sets the name to use for the command.
func WithXargsCommandName(val string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.CommandName = val
	}
}

// WithXargsExample creates an Option that sets the example to use for the command.
func WithXargsExample(val string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.Example = val
	}
}

// WithXargsLabelFilters creates an Option that sets flag name to label pairs to use as list filters.
func WithXargsLabelFilters(val map[string]string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.LabelFilters = val
	}
}

// WithXargsLabelRequirements creates an Option that sets label requirements to filter resources.
func WithXargsLabelRequirements(val []labels.Requirement) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.LabelRequirements = val
	}
}

// WithXargsLong creates an Option that sets the long description to use for the command.
func WithXargsLong(val string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.Long = val
	}
}

// WithXargsPluralFriendlyName creates an Option that sets the plural object name to display for this resource.
func WithXargsPluralFriendlyName(val string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.PluralFriendlyName = val
	}
}

// WithXargsShort creates an Option that sets the short description to use for the command.
func WithXargsShort(val string) XargsOption {
	return func(cfg *xargsConfig) {
		cfg.Short = val
	}
}

// XargsOptionDefaults gets the default values for Xargs.
func XargsOptionDefaults() XargsOptions {
	return XargsOptions{}
}
