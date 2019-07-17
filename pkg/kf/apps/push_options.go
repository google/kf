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

import (
	"io"
	"os"
)

type pushConfig struct {
	// Buildpack is skip the detect buildpack step and use the given name
	Buildpack string
	// ContainerImage is the container to deploy
	ContainerImage string
	// ContainerRegistry is the container registry's URL
	ContainerRegistry string
	// EnvironmentVariables is set environment variables
	EnvironmentVariables map[string]string
	// Grpc is setup the ports for the container to allow gRPC to work
	Grpc bool
	// MaxScale is the upper scale bound
	MaxScale int
	// MinScale is the lower scale bound
	MinScale int
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// ServiceAccount is the service account to authenticate with
	ServiceAccount string
	// SourceImage is the source code as a container image
	SourceImage string
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

// ContainerImage returns the last set value for ContainerImage or the empty value
// if not set.
func (opts PushOptions) ContainerImage() string {
	return opts.toConfig().ContainerImage
}

// ContainerRegistry returns the last set value for ContainerRegistry or the empty value
// if not set.
func (opts PushOptions) ContainerRegistry() string {
	return opts.toConfig().ContainerRegistry
}

// EnvironmentVariables returns the last set value for EnvironmentVariables or the empty value
// if not set.
func (opts PushOptions) EnvironmentVariables() map[string]string {
	return opts.toConfig().EnvironmentVariables
}

// Grpc returns the last set value for Grpc or the empty value
// if not set.
func (opts PushOptions) Grpc() bool {
	return opts.toConfig().Grpc
}

// MaxScale returns the last set value for MaxScale or the empty value
// if not set.
func (opts PushOptions) MaxScale() int {
	return opts.toConfig().MaxScale
}

// MinScale returns the last set value for MinScale or the empty value
// if not set.
func (opts PushOptions) MinScale() int {
	return opts.toConfig().MinScale
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

// SourceImage returns the last set value for SourceImage or the empty value
// if not set.
func (opts PushOptions) SourceImage() string {
	return opts.toConfig().SourceImage
}

// WithPushBuildpack creates an Option that sets skip the detect buildpack step and use the given name
func WithPushBuildpack(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Buildpack = val
	}
}

// WithPushContainerImage creates an Option that sets the container to deploy
func WithPushContainerImage(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerImage = val
	}
}

// WithPushContainerRegistry creates an Option that sets the container registry's URL
func WithPushContainerRegistry(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerRegistry = val
	}
}

// WithPushEnvironmentVariables creates an Option that sets set environment variables
func WithPushEnvironmentVariables(val map[string]string) PushOption {
	return func(cfg *pushConfig) {
		cfg.EnvironmentVariables = val
	}
}

// WithPushGrpc creates an Option that sets setup the ports for the container to allow gRPC to work
func WithPushGrpc(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.Grpc = val
	}
}

// WithPushMaxScale creates an Option that sets the upper scale bound
func WithPushMaxScale(val int) PushOption {
	return func(cfg *pushConfig) {
		cfg.MaxScale = val
	}
}

// WithPushMinScale creates an Option that sets the lower scale bound
func WithPushMinScale(val int) PushOption {
	return func(cfg *pushConfig) {
		cfg.MinScale = val
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

// WithPushSourceImage creates an Option that sets the source code as a container image
func WithPushSourceImage(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.SourceImage = val
	}
}

// PushOptionDefaults gets the default values for Push.
func PushOptionDefaults() PushOptions {
	return PushOptions{
		WithPushNamespace("default"),
		WithPushOutput(os.Stdout),
	}
}

type deployConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// DeployOption is a single option for configuring a deployConfig
type DeployOption func(*deployConfig)

// DeployOptions is a configuration set defining a deployConfig
type DeployOptions []DeployOption

// toConfig applies all the options to a new deployConfig and returns it.
func (opts DeployOptions) toConfig() deployConfig {
	cfg := deployConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeployOptions with the contents of other overriding
// the values set in this DeployOptions.
func (opts DeployOptions) Extend(other DeployOptions) DeployOptions {
	var out DeployOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeployOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeployNamespace creates an Option that sets the Kubernetes namespace to use
func WithDeployNamespace(val string) DeployOption {
	return func(cfg *deployConfig) {
		cfg.Namespace = val
	}
}

// DeployOptionDefaults gets the default values for Deploy.
func DeployOptionDefaults() DeployOptions {
	return DeployOptions{
		WithDeployNamespace("default"),
	}
}
