// This file was generated with option-builder.go, DO NOT EDIT IT.

package services

type getServiceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// GetServiceOption is a single option for configuring a getServiceConfig
type GetServiceOption func(*getServiceConfig)

// GetServiceOptions is a configuration set defining a getServiceConfig
type GetServiceOptions []GetServiceOption

// toConfig applies all the options to a new getServiceConfig and returns it.
func (opts GetServiceOptions) toConfig() getServiceConfig {
	cfg := getServiceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new GetServiceOptions with the contents of other overriding
// the values set in this GetServiceOptions.
func (opts GetServiceOptions) Extend(other GetServiceOptions) GetServiceOptions {
	var out GetServiceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts GetServiceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithGetServiceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithGetServiceNamespace(val string) GetServiceOption {
	return func(cfg *getServiceConfig) {
		cfg.Namespace = val
	}
}

// GetServiceOptionDefaults gets the default values for GetService.
func GetServiceOptionDefaults() GetServiceOptions {
	return GetServiceOptions{
		WithGetServiceNamespace("default"),
	}
}
