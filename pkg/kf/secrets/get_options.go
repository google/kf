// This file was generated with option-builder.go, DO NOT EDIT IT.

package secrets

type getConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// GetOption is a single option for configuring a getConfig
type GetOption func(*getConfig)

// GetOptions is a configuration set defining a getConfig
type GetOptions []GetOption

// toConfig applies all the options to a new getConfig and returns it.
func (opts GetOptions) toConfig() getConfig {
	cfg := getConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new GetOptions with the contents of other overriding
// the values set in this GetOptions.
func (opts GetOptions) Extend(other GetOptions) GetOptions {
	var out GetOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts GetOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithGetNamespace creates an Option that sets the Kubernetes namespace to use
func WithGetNamespace(val string) GetOption {
	return func(cfg *getConfig) {
		cfg.Namespace = val
	}
}

// GetOptionDefaults gets the default values for Get.
func GetOptionDefaults() GetOptions {
	return GetOptions{
		WithGetNamespace("default"),
	}
}
