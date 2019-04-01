// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

import (
	"io"
	"os"
)

type pushConfig struct {
	// ContainerRegistry is the container registry's URL
	ContainerRegistry string
	// DockerImage is the docker image to serve
	DockerImage string
	// EnvironmentVariables is Set environment variables.
	EnvironmentVariables []string
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
	// Path is the path of the directory to push
	Path string
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

// ContainerRegistry returns the last set value for ContainerRegistry or the empty value
// if not set.
func (opts PushOptions) ContainerRegistry() string {
	return opts.toConfig().ContainerRegistry
}

// DockerImage returns the last set value for DockerImage or the empty value
// if not set.
func (opts PushOptions) DockerImage() string {
	return opts.toConfig().DockerImage
}

// EnvironmentVariables returns the last set value for EnvironmentVariables or the empty value
// if not set.
func (opts PushOptions) EnvironmentVariables() []string {
	return opts.toConfig().EnvironmentVariables
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

// Path returns the last set value for Path or the empty value
// if not set.
func (opts PushOptions) Path() string {
	return opts.toConfig().Path
}

// ServiceAccount returns the last set value for ServiceAccount or the empty value
// if not set.
func (opts PushOptions) ServiceAccount() string {
	return opts.toConfig().ServiceAccount
}

// WithPushContainerRegistry creates an Option that sets the container registry's URL
func WithPushContainerRegistry(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerRegistry = val
	}
}

// WithPushDockerImage creates an Option that sets the docker image to serve
func WithPushDockerImage(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.DockerImage = val
	}
}

// WithPushEnvironmentVariables creates an Option that sets Set environment variables.
func WithPushEnvironmentVariables(val []string) PushOption {
	return func(cfg *pushConfig) {
		cfg.EnvironmentVariables = val
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

// WithPushPath creates an Option that sets the path of the directory to push
func WithPushPath(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Path = val
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
