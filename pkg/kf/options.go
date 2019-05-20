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

import (
	"io"
	"os"
)

type pushConfig struct {
	// Buildpack is skip the detect buildpack step and use the given name
	Buildpack string
	// ContainerRegistry is the container registry's URL
	ContainerRegistry string
	// EnvironmentVariables is set environment variables
	EnvironmentVariables []string
	// Grpc is setup the ports for the container to allow gRPC to work.
	Grpc bool
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// ServiceAccount is the service account to authenticate with
	ServiceAccount string
}

// PushOption is a single option for configuring a pushConfig
type PushOption func(*pushConfig)

// PushOptions is a configuration set defining a pushConfig
type PushOptions []PushOption

// toConfig applies all the options to a new pushConfig and returns it.
func (opts PushOptions) toConfig() pushConfig {
	cfg := pushConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new PushOptions with the contents of other overriding
// the values set in this PushOptions.
func (opts PushOptions) Extend(other PushOptions) PushOptions {
	var out PushOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Buildpack returns the last set value for Buildpack or the empty value
// if not set.
func (opts PushOptions) Buildpack() string {
	return opts.toConfig().Buildpack
}

// ContainerRegistry returns the last set value for ContainerRegistry or the empty value
// if not set.
func (opts PushOptions) ContainerRegistry() string {
	return opts.toConfig().ContainerRegistry
}

// EnvironmentVariables returns the last set value for EnvironmentVariables or the empty value
// if not set.
func (opts PushOptions) EnvironmentVariables() []string {
	return opts.toConfig().EnvironmentVariables
}

// Grpc returns the last set value for Grpc or the empty value
// if not set.
func (opts PushOptions) Grpc() bool {
	return opts.toConfig().Grpc
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts PushOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Output returns the last set value for Output or the empty value
// if not set.
func (opts PushOptions) Output() io.Writer {
	return opts.toConfig().Output
}

// ServiceAccount returns the last set value for ServiceAccount or the empty value
// if not set.
func (opts PushOptions) ServiceAccount() string {
	return opts.toConfig().ServiceAccount
}

// WithPushBuildpack creates an Option that sets skip the detect buildpack step and use the given name
func WithPushBuildpack(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Buildpack = val
	}
}

// WithPushContainerRegistry creates an Option that sets the container registry's URL
func WithPushContainerRegistry(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerRegistry = val
	}
}

// WithPushEnvironmentVariables creates an Option that sets set environment variables
func WithPushEnvironmentVariables(val []string) PushOption {
	return func(cfg *pushConfig) {
		cfg.EnvironmentVariables = val
	}
}

// WithPushGrpc creates an Option that sets setup the ports for the container to allow gRPC to work.
func WithPushGrpc(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.Grpc = val
	}
}

// WithPushNamespace creates an Option that sets the Kubernetes namespace to use
func WithPushNamespace(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Namespace = val
	}
}

// WithPushOutput creates an Option that sets the io.Writer to write output such as build logs
func WithPushOutput(val io.Writer) PushOption {
	return func(cfg *pushConfig) {
		cfg.Output = val
	}
}

// WithPushServiceAccount creates an Option that sets the service account to authenticate with
func WithPushServiceAccount(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ServiceAccount = val
	}
}

// PushOptionDefaults gets the default values for Push.
func PushOptionDefaults() PushOptions {
	return PushOptions{
		WithPushNamespace("default"),
		WithPushOutput(os.Stdout),
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

type listConfig struct {
	// AppName is the specific app name to look for
	AppName string
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

// AppName returns the last set value for AppName or the empty value
// if not set.
func (opts ListOptions) AppName() string {
	return opts.toConfig().AppName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListAppName creates an Option that sets the specific app name to look for
func WithListAppName(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.AppName = val
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

type listConfigurationsConfig struct {
	// AppName is the specific app name to look for
	AppName string
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// ListConfigurationsOption is a single option for configuring a listConfigurationsConfig
type ListConfigurationsOption func(*listConfigurationsConfig)

// ListConfigurationsOptions is a configuration set defining a listConfigurationsConfig
type ListConfigurationsOptions []ListConfigurationsOption

// toConfig applies all the options to a new listConfigurationsConfig and returns it.
func (opts ListConfigurationsOptions) toConfig() listConfigurationsConfig {
	cfg := listConfigurationsConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListConfigurationsOptions with the contents of other overriding
// the values set in this ListConfigurationsOptions.
func (opts ListConfigurationsOptions) Extend(other ListConfigurationsOptions) ListConfigurationsOptions {
	var out ListConfigurationsOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// AppName returns the last set value for AppName or the empty value
// if not set.
func (opts ListConfigurationsOptions) AppName() string {
	return opts.toConfig().AppName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListConfigurationsOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListConfigurationsAppName creates an Option that sets the specific app name to look for
func WithListConfigurationsAppName(val string) ListConfigurationsOption {
	return func(cfg *listConfigurationsConfig) {
		cfg.AppName = val
	}
}

// WithListConfigurationsNamespace creates an Option that sets the Kubernetes namespace to use
func WithListConfigurationsNamespace(val string) ListConfigurationsOption {
	return func(cfg *listConfigurationsConfig) {
		cfg.Namespace = val
	}
}

// ListConfigurationsOptionDefaults gets the default values for ListConfigurations.
func ListConfigurationsOptionDefaults() ListConfigurationsOptions {
	return ListConfigurationsOptions{
		WithListConfigurationsNamespace("default"),
	}
}

type listEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// ListEnvOption is a single option for configuring a listEnvConfig
type ListEnvOption func(*listEnvConfig)

// ListEnvOptions is a configuration set defining a listEnvConfig
type ListEnvOptions []ListEnvOption

// toConfig applies all the options to a new listEnvConfig and returns it.
func (opts ListEnvOptions) toConfig() listEnvConfig {
	cfg := listEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListEnvOptions with the contents of other overriding
// the values set in this ListEnvOptions.
func (opts ListEnvOptions) Extend(other ListEnvOptions) ListEnvOptions {
	var out ListEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithListEnvNamespace(val string) ListEnvOption {
	return func(cfg *listEnvConfig) {
		cfg.Namespace = val
	}
}

// ListEnvOptionDefaults gets the default values for ListEnv.
func ListEnvOptionDefaults() ListEnvOptions {
	return ListEnvOptions{
		WithListEnvNamespace("default"),
	}
}

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
