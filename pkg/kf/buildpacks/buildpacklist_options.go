// This file was generated with option-builder.go, DO NOT EDIT IT.

package buildpacks

type buildpackListConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// BuildpackListOption is a single option for configuring a buildpackListConfig
type BuildpackListOption func(*buildpackListConfig)

// BuildpackListOptions is a configuration set defining a buildpackListConfig
type BuildpackListOptions []BuildpackListOption

// toConfig applies all the options to a new buildpackListConfig and returns it.
func (opts BuildpackListOptions) toConfig() buildpackListConfig {
	cfg := buildpackListConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new BuildpackListOptions with the contents of other overriding
// the values set in this BuildpackListOptions.
func (opts BuildpackListOptions) Extend(other BuildpackListOptions) BuildpackListOptions {
	var out BuildpackListOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts BuildpackListOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithBuildpackListNamespace creates an Option that sets the Kubernetes namespace to use
func WithBuildpackListNamespace(val string) BuildpackListOption {
	return func(cfg *buildpackListConfig) {
		cfg.Namespace = val
	}
}

// BuildpackListOptionDefaults gets the default values for BuildpackList.
func BuildpackListOptionDefaults() BuildpackListOptions {
	return BuildpackListOptions{
		WithBuildpackListNamespace("default"),
	}
}
