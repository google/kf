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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type pushConfig struct {
	// AppSpecInstances is Scaling information for the service
	AppSpecInstances v1alpha1.AppSpecInstances
	// Args is the app container arguments
	Args []string
	// Buildpack is skip the detect buildpack step and use the given name
	Buildpack string
	// Command is the app container entrypoint
	Command []string
	// ContainerImage is the container to deploy
	ContainerImage string
	// DefaultRouteDomain is Domain for a defaultroute. Only used if a route doesn't already exist
	DefaultRouteDomain string
	// EnvironmentVariables is set environment variables
	EnvironmentVariables map[string]string
	// Grpc is setup the ports for the container to allow gRPC to work
	Grpc bool
	// HealthCheck is the health check to use on the app
	HealthCheck *corev1.Probe
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// RandomRouteDomain is Domain for a random route. Only used if a route doesn't already exist
	RandomRouteDomain string
	// ResourceRequests is Resource requests for the container
	ResourceRequests corev1.ResourceList
	// Routes is routes for the app
	Routes []v1alpha1.RouteSpecFields
	// ServiceBindings is a list of Services to bind to the app
	ServiceBindings []v1alpha1.AppSpecServiceBinding
	// SourceImage is the source code as a container image
	SourceImage string
	// Stack is the builder stack to use for buildpack based apps
	Stack string
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

// AppSpecInstances returns the last set value for AppSpecInstances or the empty value
// if not set.
func (opts PushOptions) AppSpecInstances() v1alpha1.AppSpecInstances {
	return opts.toConfig().AppSpecInstances
}

// Args returns the last set value for Args or the empty value
// if not set.
func (opts PushOptions) Args() []string {
	return opts.toConfig().Args
}

// Buildpack returns the last set value for Buildpack or the empty value
// if not set.
func (opts PushOptions) Buildpack() string {
	return opts.toConfig().Buildpack
}

// Command returns the last set value for Command or the empty value
// if not set.
func (opts PushOptions) Command() []string {
	return opts.toConfig().Command
}

// ContainerImage returns the last set value for ContainerImage or the empty value
// if not set.
func (opts PushOptions) ContainerImage() string {
	return opts.toConfig().ContainerImage
}

// DefaultRouteDomain returns the last set value for DefaultRouteDomain or the empty value
// if not set.
func (opts PushOptions) DefaultRouteDomain() string {
	return opts.toConfig().DefaultRouteDomain
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

// RandomRouteDomain returns the last set value for RandomRouteDomain or the empty value
// if not set.
func (opts PushOptions) RandomRouteDomain() string {
	return opts.toConfig().RandomRouteDomain
}

// ResourceRequests returns the last set value for ResourceRequests or the empty value
// if not set.
func (opts PushOptions) ResourceRequests() corev1.ResourceList {
	return opts.toConfig().ResourceRequests
}

// Routes returns the last set value for Routes or the empty value
// if not set.
func (opts PushOptions) Routes() []v1alpha1.RouteSpecFields {
	return opts.toConfig().Routes
}

// ServiceBindings returns the last set value for ServiceBindings or the empty value
// if not set.
func (opts PushOptions) ServiceBindings() []v1alpha1.AppSpecServiceBinding {
	return opts.toConfig().ServiceBindings
}

// SourceImage returns the last set value for SourceImage or the empty value
// if not set.
func (opts PushOptions) SourceImage() string {
	return opts.toConfig().SourceImage
}

// Stack returns the last set value for Stack or the empty value
// if not set.
func (opts PushOptions) Stack() string {
	return opts.toConfig().Stack
}

// WithPushAppSpecInstances creates an Option that sets Scaling information for the service
func WithPushAppSpecInstances(val v1alpha1.AppSpecInstances) PushOption {
	return func(cfg *pushConfig) {
		cfg.AppSpecInstances = val
	}
}

// WithPushArgs creates an Option that sets the app container arguments
func WithPushArgs(val []string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Args = val
	}
}

// WithPushBuildpack creates an Option that sets skip the detect buildpack step and use the given name
func WithPushBuildpack(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Buildpack = val
	}
}

// WithPushCommand creates an Option that sets the app container entrypoint
func WithPushCommand(val []string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Command = val
	}
}

// WithPushContainerImage creates an Option that sets the container to deploy
func WithPushContainerImage(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerImage = val
	}
}

// WithPushDefaultRouteDomain creates an Option that sets Domain for a defaultroute. Only used if a route doesn't already exist
func WithPushDefaultRouteDomain(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.DefaultRouteDomain = val
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

// WithPushRandomRouteDomain creates an Option that sets Domain for a random route. Only used if a route doesn't already exist
func WithPushRandomRouteDomain(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.RandomRouteDomain = val
	}
}

// WithPushResourceRequests creates an Option that sets Resource requests for the container
func WithPushResourceRequests(val corev1.ResourceList) PushOption {
	return func(cfg *pushConfig) {
		cfg.ResourceRequests = val
	}
}

// WithPushRoutes creates an Option that sets routes for the app
func WithPushRoutes(val []v1alpha1.RouteSpecFields) PushOption {
	return func(cfg *pushConfig) {
		cfg.Routes = val
	}
}

// WithPushServiceBindings creates an Option that sets a list of Services to bind to the app
func WithPushServiceBindings(val []v1alpha1.AppSpecServiceBinding) PushOption {
	return func(cfg *pushConfig) {
		cfg.ServiceBindings = val
	}
}

// WithPushSourceImage creates an Option that sets the source code as a container image
func WithPushSourceImage(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.SourceImage = val
	}
}

// WithPushStack creates an Option that sets the builder stack to use for buildpack based apps
func WithPushStack(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Stack = val
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
