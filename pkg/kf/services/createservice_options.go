// This file was generated with option-builder.go, DO NOT EDIT IT.

package services

type createServiceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// Params is service-specific configuration parameters.
	Params map[string]interface{}
}

// CreateServiceOption is a single option for configuring a createServiceConfig
type CreateServiceOption func(*createServiceConfig)

// CreateServiceOptions is a configuration set defining a createServiceConfig
type CreateServiceOptions []CreateServiceOption

// toConfig applies all the options to a new createServiceConfig and returns it.
func (opts CreateServiceOptions) toConfig() createServiceConfig {
	cfg := createServiceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new CreateServiceOptions with the contents of other overriding
// the values set in this CreateServiceOptions.
func (opts CreateServiceOptions) Extend(other CreateServiceOptions) CreateServiceOptions {
	var out CreateServiceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateServiceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Params returns the last set value for Params or the empty value
// if not set.
func (opts CreateServiceOptions) Params() map[string]interface{} {
	return opts.toConfig().Params
}

// WithCreateServiceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithCreateServiceNamespace(val string) CreateServiceOption {
	return func(cfg *createServiceConfig) {
		cfg.Namespace = val
	}
}

// WithCreateServiceParams creates an Option that sets service-specific configuration parameters.
func WithCreateServiceParams(val map[string]interface{}) CreateServiceOption {
	return func(cfg *createServiceConfig) {
		cfg.Params = val
	}
}

// CreateServiceOptionDefaults gets the default values for CreateService.
func CreateServiceOptionDefaults() CreateServiceOptions {
	return CreateServiceOptions{
		WithCreateServiceNamespace("default"),
	}
}
