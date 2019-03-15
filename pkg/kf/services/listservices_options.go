// This file was generated with option-builder.go, DO NOT EDIT IT.

package services

type listServicesConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// ListServicesOption is a single option for configuring a listServicesConfig
type ListServicesOption func(*listServicesConfig)

// ListServicesOptions is a configuration set defining a listServicesConfig
type ListServicesOptions []ListServicesOption

// toConfig applies all the options to a new listServicesConfig and returns it.
func (opts ListServicesOptions) toConfig() listServicesConfig {
	cfg := listServicesConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new ListServicesOptions with the contents of other overriding
// the values set in this ListServicesOptions.
func (opts ListServicesOptions) Extend(other ListServicesOptions) ListServicesOptions {
	var out ListServicesOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListServicesOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithListServicesNamespace creates an Option that sets the Kubernetes namespace to use.
func WithListServicesNamespace(val string) ListServicesOption {
	return func(cfg *listServicesConfig) {
		cfg.Namespace = val
	}
}

// ListServicesOptionDefaults gets the default values for ListServices.
func ListServicesOptionDefaults() ListServicesOptions {
	return ListServicesOptions{
		WithListServicesNamespace("default"),
	}
}
