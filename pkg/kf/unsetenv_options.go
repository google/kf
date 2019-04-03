// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf

type unsetEnvConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// UnsetEnvOption is a single option for configuring a unsetEnvConfig
type UnsetEnvOption func(*unsetEnvConfig)

// UnsetEnvOptions is a configuration set defining a unsetEnvConfig
type UnsetEnvOptions []UnsetEnvOption

// toConfig applies all the options to a new unsetEnvConfig and returns it.
func (opts UnsetEnvOptions) toConfig() unsetEnvConfig {
	cfg := unsetEnvConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new UnsetEnvOptions with the contents of other overriding
// the values set in this UnsetEnvOptions.
func (opts UnsetEnvOptions) Extend(other UnsetEnvOptions) UnsetEnvOptions {
	var out UnsetEnvOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts UnsetEnvOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithUnsetEnvNamespace creates an Option that sets the Kubernetes namespace to use
func WithUnsetEnvNamespace(val string) UnsetEnvOption {
	return func(cfg *unsetEnvConfig) {
		cfg.Namespace = val
	}
}

// UnsetEnvOptionDefaults gets the default values for UnsetEnv.
func UnsetEnvOptionDefaults() UnsetEnvOptions {
	return UnsetEnvOptions{
		WithUnsetEnvNamespace("default"),
	}
}
