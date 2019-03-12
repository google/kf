// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

import (
	"io"
)

type pushConfig struct { // Namespace is the Kubertes namespace to use
	Namespace string
	// Path is the path of the directory to push
	Path string
	// ContainerRegistry is the container registry's URL
	ContainerRegistry string
	// ServiceAccount is the service account to authenticate with
	ServiceAccount string
	// Output is the io.Writer to write output such as build logs
	Output io.Writer
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

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts PushOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Path returns the last set value for Path or the empty value
// if not set.
func (opts PushOptions) Path() string {
	return opts.toConfig().Path
}

// ContainerRegistry returns the last set value for ContainerRegistry or the empty value
// if not set.
func (opts PushOptions) ContainerRegistry() string {
	return opts.toConfig().ContainerRegistry
}

// ServiceAccount returns the last set value for ServiceAccount or the empty value
// if not set.
func (opts PushOptions) ServiceAccount() string {
	return opts.toConfig().ServiceAccount
}

// Output returns the last set value for Output or the empty value
// if not set.
func (opts PushOptions) Output() io.Writer {
	return opts.toConfig().Output
}

// WithPushNamespace creates an Option that sets the Kubertes namespace to use
func WithPushNamespace(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Namespace = val
	}
}

// WithPushPath creates an Option that sets the path of the directory to push
func WithPushPath(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.Path = val
	}
}

// WithPushContainerRegistry creates an Option that sets the container registry's URL
func WithPushContainerRegistry(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ContainerRegistry = val
	}
}

// WithPushServiceAccount creates an Option that sets the service account to authenticate with
func WithPushServiceAccount(val string) PushOption {
	return func(cfg *pushConfig) {
		cfg.ServiceAccount = val
	}
}

// WithPushOutput creates an Option that sets the io.Writer to write output such as build logs
func WithPushOutput(val io.Writer) PushOption {
	return func(cfg *pushConfig) {
		cfg.Output = val
	}
}
