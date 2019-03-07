// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

type listConfig struct { // Namespace is the Kubertes namespace to use
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

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListNamespace creates an Option that sets the Kubertes namespace to use
func WithListNamespace(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Namespace = val
	}
}
