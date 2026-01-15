// Copyright 2026 Google LLC
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
	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"io"
	corev1 "k8s.io/api/core/v1"
	"os"
)

type pushConfig struct {
	// ADXBuild is use AppDevExperience for builds
	ADXBuild bool
	// ADXContainerRegistry is the container registry configured on the Space
	ADXContainerRegistry string
	// ADXDockerfile is the path to the dockerfile to us with the AppDevExperience build
	ADXDockerfile string
	// ADXStack is the stack to use with the AppDevExperience build
	ADXStack config.StackV3Definition
	// Annotations is Annotations to add to the pushed app.
	Annotations map[string]string
	// AppSpecInstances is Scaling information for the service
	AppSpecInstances v1alpha1.AppSpecInstances
	// Build is a custom Tekton task used for the build
	Build *v1alpha1.BuildSpec
	// Container is the app container template
	Container corev1.Container
	// ContainerImage is the container to deploy
	ContainerImage *string
	// GenerateDefaultRoute is returns true if the app should receive a default route if a route does not already exist
	GenerateDefaultRoute bool
	// GenerateRandomRoute is returns true if the app should receive a random route if a route doesn't already exist
	GenerateRandomRoute bool
	// Labels is Labels to add to the pushed app.
	Labels map[string]string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// Routes is routes for the app
	Routes []v1alpha1.RouteWeightBinding
	// ServiceBindings is a list of Services to bind to the app
	ServiceBindings []v1alpha1.ServiceInstanceBinding
	// SourcePath is the path to the source code directory
	SourcePath string
	// Space is the Space to use
	Space string
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

// ADXBuild returns the last set value for ADXBuild or the empty value
// if not set.
func (opts PushOptions) ADXBuild() bool {
	return opts.toConfig().ADXBuild
}

// ADXContainerRegistry returns the last set value for ADXContainerRegistry or the empty value
// if not set.
func (opts PushOptions) ADXContainerRegistry() string {
	return opts.toConfig().ADXContainerRegistry
}

// ADXDockerfile returns the last set value for ADXDockerfile or the empty value
// if not set.
func (opts PushOptions) ADXDockerfile() string {
	return opts.toConfig().ADXDockerfile
}

// ADXStack returns the last set value for ADXStack or the empty value
// if not set.
func (opts PushOptions) ADXStack() config.StackV3Definition {
	return opts.toConfig().ADXStack
}

// Annotations returns the last set value for Annotations or the empty value
// if not set.
func (opts PushOptions) Annotations() map[string]string {
	return opts.toConfig().Annotations
}

// AppSpecInstances returns the last set value for AppSpecInstances or the empty value
// if not set.
func (opts PushOptions) AppSpecInstances() v1alpha1.AppSpecInstances {
	return opts.toConfig().AppSpecInstances
}

// Build returns the last set value for Build or the empty value
// if not set.
func (opts PushOptions) Build() *v1alpha1.BuildSpec {
	return opts.toConfig().Build
}

// Container returns the last set value for Container or the empty value
// if not set.
func (opts PushOptions) Container() corev1.Container {
	return opts.toConfig().Container
}

// ContainerImage returns the last set value for ContainerImage or the empty value
// if not set.
func (opts PushOptions) ContainerImage() *string {
	return opts.toConfig().ContainerImage
}

// GenerateDefaultRoute returns the last set value for GenerateDefaultRoute or the empty value
// if not set.
func (opts PushOptions) GenerateDefaultRoute() bool {
	return opts.toConfig().GenerateDefaultRoute
}

// GenerateRandomRoute returns the last set value for GenerateRandomRoute or the empty value
// if not set.
func (opts PushOptions) GenerateRandomRoute() bool {
	return opts.toConfig().GenerateRandomRoute
}

// Labels returns the last set value for Labels or the empty value
// if not set.
func (opts PushOptions) Labels() map[string]string {
	return opts.toConfig().Labels
}

// Output returns the last set value for Output or the empty value
// if not set.
func (opts PushOptions) Output() io.Writer {
	return opts.toConfig().Output
}

// Routes returns the last set value for Routes or the empty value
// if not set.
func (opts PushOptions) Routes() []v1alpha1.RouteWeightBinding {
	return opts.toConfig().Routes
}

// ServiceBindings returns the last set value for ServiceBindings or the empty value
// if not set.
func (opts PushOptions) ServiceBindings() []v1alpha1.ServiceInstanceBinding {
	return opts.toConfig().ServiceBindings
}

// SourcePath returns the last set value for SourcePath or the empty value
// if not set.
func (opts PushOptions) SourcePath() string {
	return opts.toConfig().SourcePath
}

// Space returns the last set value for Space or the empty value
// if not set.
func (opts PushOptions) Space() string {
	return opts.toConfig().Space
}

// WithPushADXBuild creates an Option that sets use AppDevExperience for builds
func WithPushADXBuild(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.ADXBuild = val
	}
}

// WithPushADXContainerRegistry creates an Option that sets the container registry configured on the Space
func WithPushADXContainerRegistry(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ADXContainerRegistry = val
	}
}

// WithPushADXDockerfile creates an Option that sets the path to the dockerfile to us with the AppDevExperience build
func WithPushADXDockerfile(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ADXDockerfile = val
	}
}

// WithPushADXStack creates an Option that sets the stack to use with the AppDevExperience build
func WithPushADXStack(val config.StackV3Definition) PushOption {
	return func(cfg *pushConfig) {
		cfg.ADXStack = val
	}
}

// WithPushAnnotations creates an Option that sets Annotations to add to the pushed app.
func WithPushAnnotations(val map[string]string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Annotations = val
	}
}

// WithPushAppSpecInstances creates an Option that sets Scaling information for the service
func WithPushAppSpecInstances(val v1alpha1.AppSpecInstances) PushOption {
	return func(cfg *pushConfig) {
		cfg.AppSpecInstances = val
	}
}

// WithPushBuild creates an Option that sets a custom Tekton task used for the build
func WithPushBuild(val *v1alpha1.BuildSpec) PushOption {
	return func(cfg *pushConfig) {
		cfg.Build = val
	}
}

// WithPushContainer creates an Option that sets the app container template
func WithPushContainer(val corev1.Container) PushOption {
	return func(cfg *pushConfig) {
		cfg.Container = val
	}
}

// WithPushContainerImage creates an Option that sets the container to deploy
func WithPushContainerImage(val *string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerImage = val
	}
}

// WithPushGenerateDefaultRoute creates an Option that sets returns true if the app should receive a default route if a route does not already exist
func WithPushGenerateDefaultRoute(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.GenerateDefaultRoute = val
	}
}

// WithPushGenerateRandomRoute creates an Option that sets returns true if the app should receive a random route if a route doesn't already exist
func WithPushGenerateRandomRoute(val bool) PushOption {
	return func(cfg *pushConfig) {
		cfg.GenerateRandomRoute = val
	}
}

// WithPushLabels creates an Option that sets Labels to add to the pushed app.
func WithPushLabels(val map[string]string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Labels = val
	}
}

// WithPushOutput creates an Option that sets the io.Writer to write output such as build logs
func WithPushOutput(val io.Writer) PushOption {
	return func(cfg *pushConfig) {
		cfg.Output = val
	}
}

// WithPushRoutes creates an Option that sets routes for the app
func WithPushRoutes(val []v1alpha1.RouteWeightBinding) PushOption {
	return func(cfg *pushConfig) {
		cfg.Routes = val
	}
}

// WithPushServiceBindings creates an Option that sets a list of Services to bind to the app
func WithPushServiceBindings(val []v1alpha1.ServiceInstanceBinding) PushOption {
	return func(cfg *pushConfig) {
		cfg.ServiceBindings = val
	}
}

// WithPushSourcePath creates an Option that sets the path to the source code directory
func WithPushSourcePath(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.SourcePath = val
	}
}

// WithPushSpace creates an Option that sets the Space to use
func WithPushSpace(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Space = val
	}
}

// PushOptionDefaults gets the default values for Push.
func PushOptionDefaults() PushOptions {
	return PushOptions{
		WithPushOutput(os.Stdout),
		WithPushSpace("default"),
	}
}
