// This file was generated with option-builder.go, DO NOT EDIT IT.

package servicebindings

type deleteConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// DeleteOption is a single option for configuring a deleteConfig
type DeleteOption func(*deleteConfig)

// DeleteOptions is a configuration set defining a deleteConfig
type DeleteOptions []DeleteOption

// toConfig applies all the options to a new deleteConfig and returns it.
func (opts DeleteOptions) toConfig() deleteConfig {
	cfg := deleteConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteOptions with the contents of other overriding
// the values set in this DeleteOptions.
func (opts DeleteOptions) Extend(other DeleteOptions) DeleteOptions {
	var out DeleteOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeleteOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeleteNamespace creates an Option that sets the Kubernetes namespace to use.
func WithDeleteNamespace(val string) DeleteOption {
	return func(cfg *deleteConfig) {
		cfg.Namespace = val
	}
}

// DeleteOptionDefaults gets the default values for Delete.
func DeleteOptionDefaults() DeleteOptions {
	return DeleteOptions{
		WithDeleteNamespace("default"),
	}
}
