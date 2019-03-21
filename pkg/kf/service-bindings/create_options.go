// This file was generated with option-builder.go, DO NOT EDIT IT.

package servicebindings

type createConfig struct {
	// BindingName is name to expose service instance to app process with.
	BindingName string
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// Params is service-specific configuration parameters.
	Params map[string]interface{}
}

// CreateOption is a single option for configuring a createConfig
type CreateOption func(*createConfig)

// CreateOptions is a configuration set defining a createConfig
type CreateOptions []CreateOption

// toConfig applies all the options to a new createConfig and returns it.
func (opts CreateOptions) toConfig() createConfig {
	cfg := createConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new CreateOptions with the contents of other overriding
// the values set in this CreateOptions.
func (opts CreateOptions) Extend(other CreateOptions) CreateOptions {
	var out CreateOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// BindingName returns the last set value for BindingName or the empty value
// if not set.
func (opts CreateOptions) BindingName() string {
	return opts.toConfig().BindingName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Params returns the last set value for Params or the empty value
// if not set.
func (opts CreateOptions) Params() map[string]interface{} {
	return opts.toConfig().Params
}

// WithCreateBindingName creates an Option that sets name to expose service instance to app process with.
func WithCreateBindingName(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.BindingName = val
	}
}

// WithCreateNamespace creates an Option that sets the Kubernetes namespace to use.
func WithCreateNamespace(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Namespace = val
	}
}

// WithCreateParams creates an Option that sets service-specific configuration parameters.
func WithCreateParams(val map[string]interface{}) CreateOption {
	return func(cfg *createConfig) {
		cfg.Params = val
	}
}

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{
		WithCreateNamespace("default"),
	}
}
