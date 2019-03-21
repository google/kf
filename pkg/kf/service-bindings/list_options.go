// This file was generated with option-builder.go, DO NOT EDIT IT.

package servicebindings

type listConfig struct {
	// AppName is filter the results to bindings for the given app.
	AppName string
	// Namespace is the Kubernetes namespace to use.
	Namespace string
	// ServiceInstance is filter the results to bindings for the given service instance.
	ServiceInstance string
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

// Extend creates a new ListOptions with the contents of other overriding
// the values set in this ListOptions.
func (opts ListOptions) Extend(other ListOptions) ListOptions {
	var out ListOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// AppName returns the last set value for AppName or the empty value
// if not set.
func (opts ListOptions) AppName() string {
	return opts.toConfig().AppName
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts ListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// ServiceInstance returns the last set value for ServiceInstance or the empty value
// if not set.
func (opts ListOptions) ServiceInstance() string {
	return opts.toConfig().ServiceInstance
}

// WithListAppName creates an Option that sets filter the results to bindings for the given app.
func WithListAppName(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.AppName = val
	}
}

// WithListNamespace creates an Option that sets the Kubernetes namespace to use.
func WithListNamespace(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.Namespace = val
	}
}

// WithListServiceInstance creates an Option that sets filter the results to bindings for the given service instance.
func WithListServiceInstance(val string) ListOption {
	return func(cfg *listConfig) {
		cfg.ServiceInstance = val
	}
}

// ListOptionDefaults gets the default values for List.
func ListOptionDefaults() ListOptions {
	return ListOptions{
		WithListNamespace("default"),
	}
}
