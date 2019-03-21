// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

type setEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// SetEnvOption is a single option for configuring a setEnvConfig
type SetEnvOption func(*setEnvConfig)

// SetEnvOptions is a configuration set defining a setEnvConfig
type SetEnvOptions []SetEnvOption

// toConfig applies all the options to a new setEnvConfig and returns it.
func (opts SetEnvOptions) toConfig() setEnvConfig {
	cfg := setEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new SetEnvOptions with the contents of other overriding
// the values set in this SetEnvOptions.
func (opts SetEnvOptions) Extend(other SetEnvOptions) SetEnvOptions {
	var out SetEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts SetEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithSetEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithSetEnvNamespace(val string) SetEnvOption {
	return func(cfg *setEnvConfig) {
		cfg.Namespace = val
	}
}

// SetEnvOptionDefaults gets the default values for SetEnv.
func SetEnvOptionDefaults() SetEnvOptions {
	return SetEnvOptions{
		WithSetEnvNamespace("default"),
	}
}
