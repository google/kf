// This file was generated with option-builder.go, DO NOT EDIT IT.

package services

type marketplaceConfig struct {
	// Namespace is the Kubernetes namespace to use.
	Namespace string
}

// MarketplaceOption is a single option for configuring a marketplaceConfig
type MarketplaceOption func(*marketplaceConfig)

// MarketplaceOptions is a configuration set defining a marketplaceConfig
type MarketplaceOptions []MarketplaceOption

// toConfig applies all the options to a new marketplaceConfig and returns it.
func (opts MarketplaceOptions) toConfig() marketplaceConfig {
	cfg := marketplaceConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new MarketplaceOptions with the contents of other overriding
// the values set in this MarketplaceOptions.
func (opts MarketplaceOptions) Extend(other MarketplaceOptions) MarketplaceOptions {
	var out MarketplaceOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts MarketplaceOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithMarketplaceNamespace creates an Option that sets the Kubernetes namespace to use.
func WithMarketplaceNamespace(val string) MarketplaceOption {
	return func(cfg *marketplaceConfig) {
		cfg.Namespace = val
	}
}

// MarketplaceOptionDefaults gets the default values for Marketplace.
func MarketplaceOptionDefaults() MarketplaceOptions {
	return MarketplaceOptions{
		WithMarketplaceNamespace("default"),
	}
}
