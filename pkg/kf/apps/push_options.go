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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
)

type pushConfig struct {
	// Buildpack is skip the detect buildpack step and use the given name
	Buildpack string
	// CPU is app CPU request
	CPU *resource.Quantity
	// ContainerImage is the container to deploy
	ContainerImage string
	// ContainerRegistry is the container registry's URL
	ContainerRegistry string
	// DiskQuota is app disk storage quota
	DiskQuota *resource.Quantity
	// EnvironmentVariables is set environment variables
	EnvironmentVariables map[string]string
	// Grpc is setup the ports for the container to allow gRPC to work
	Grpc bool
	// HealthCheck is the health check to use on the app
	HealthCheck *corev1.Probe
	// MaxScale is the upper scale bound
	MaxScale int
	// Memory is app memory request
	Memory *resource.Quantity
	// MinScale is the lower scale bound
	MinScale int
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// NoStart is setup the app without starting it
	NoStart bool
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// Routes is routes for the app
	Routes []v1alpha1.RouteSpecFields
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

// CPU returns the last set value for CPU or the empty value
// if not set.
func (opts PushOptions) CPU() *resource.Quantity {
	return opts.toConfig().CPU
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

// DiskQuota returns the last set value for DiskQuota or the empty value
// if not set.
func (opts PushOptions) DiskQuota() *resource.Quantity {
	return opts.toConfig().DiskQuota
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

// HealthCheck returns the last set value for HealthCheck or the empty value
// if not set.
func (opts PushOptions) HealthCheck() *corev1.Probe {
	return opts.toConfig().HealthCheck
}

// MaxScale returns the last set value for MaxScale or the empty value
// if not set.
func (opts PushOptions) MaxScale() int {
	return opts.toConfig().MaxScale
}

// Memory returns the last set value for Memory or the empty value
// if not set.
func (opts PushOptions) Memory() *resource.Quantity {
	return opts.toConfig().Memory
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

// NoStart returns the last set value for NoStart or the empty value
// if not set.
func (opts PushOptions) NoStart() bool {
	return opts.toConfig().NoStart
}

// Output returns the last set value for Output or the empty value
// if not set.
func (opts PushOptions) Output() io.Writer {
	return opts.toConfig().Output
}

// Routes returns the last set value for Routes or the empty value
// if not set.
func (opts PushOptions) Routes() []v1alpha1.RouteSpecFields {
	return opts.toConfig().Routes
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

// WithPushCPU creates an Option that sets app CPU request
func WithPushCPU(val *resource.Quantity) PushOption {
	return func(cfg *pushConfig) {
		cfg.CPU = val
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

// WithPushDiskQuota creates an Option that sets app disk storage quota
func WithPushDiskQuota(val *resource.Quantity) PushOption {
	return func(cfg *pushConfig) {
		cfg.DiskQuota = val
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

// WithPushHealthCheck creates an Option that sets the health check to use on the app
func WithPushHealthCheck(val *corev1.Probe) PushOption {
	return func(cfg *pushConfig) {
		cfg.HealthCheck = val
	}
}

// WithPushMaxScale creates an Option that sets the upper scale bound
func WithPushMaxScale(val int) PushOption {
	return func(cfg *pushConfig) {
		cfg.MaxScale = val
	}
}

// WithPushMemory creates an Option that sets app memory request
func WithPushMemory(val *resource.Quantity) PushOption {
	return func(cfg *pushConfig) {
		cfg.Memory = val
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

// WithPushNoStart creates an Option that sets setup the app without starting it
func WithPushNoStart(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.NoStart = val
	}
}

// WithPushOutput creates an Option that sets the io.Writer to write output such as build logs
func WithPushOutput(val io.Writer) PushOption {
	return func(cfg *pushConfig) {
		cfg.Output = val
	}
}

// WithPushRoutes creates an Option that sets routes for the app
func WithPushRoutes(val []v1alpha1.RouteSpecFields) PushOption {
	return func(cfg *pushConfig) {
		cfg.Routes = val
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
