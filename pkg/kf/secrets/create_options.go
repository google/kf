// This file was generated with option-builder.go, DO NOT EDIT IT.

package secrets

type createConfig struct {
	// Data is data to store in the secret. Values MUST be base64.
	Data map[string][]byte
	// Labels is labels to set on the secret.
	Labels map[string]string
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// StringData is data to store in the secret. Values are encoded in base64 automatically.
	StringData map[string]string
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

// Data returns the last set value for Data or the empty value
// if not set.
func (opts CreateOptions) Data() map[string][]byte {
	return opts.toConfig().Data
}

// Labels returns the last set value for Labels or the empty value
// if not set.
func (opts CreateOptions) Labels() map[string]string {
	return opts.toConfig().Labels
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// StringData returns the last set value for StringData or the empty value
// if not set.
func (opts CreateOptions) StringData() map[string]string {
	return opts.toConfig().StringData
}

// WithCreateData creates an Option that sets data to store in the secret. Values MUST be base64.
func WithCreateData(val map[string][]byte) CreateOption {
	return func(cfg *createConfig) {
		cfg.Data = val
	}
}

// WithCreateLabels creates an Option that sets labels to set on the secret.
func WithCreateLabels(val map[string]string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Labels = val
	}
}

// WithCreateNamespace creates an Option that sets the Kubernetes namespace to use
func WithCreateNamespace(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Namespace = val
	}
}

// WithCreateStringData creates an Option that sets data to store in the secret. Values are encoded in base64 automatically.
func WithCreateStringData(val map[string]string) CreateOption {
	return func(cfg *createConfig) {
		cfg.StringData = val
	}
}

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{
		WithCreateNamespace("default"),
	}
}
