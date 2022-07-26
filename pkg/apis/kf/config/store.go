// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package config

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/configmap"
)

type cfgKey struct{}

// Config of Kf
// +k8s:deepcopy-gen=false
type Config struct {
	defaults  *DefaultsConfig
	loadError error
}

// Defaults returns the DefaultsConfig and an error if there was an issue loading the config.
func (c *Config) Defaults() (*DefaultsConfig, error) {
	return c.defaults, c.loadError
}

// CreateConfigForTest creates a *Config with a custom *DefaultsConfig. This should be used for tests only.
func CreateConfigForTest(defaults *DefaultsConfig) *Config {
	cfg := &Config{}
	cfg.defaults = defaults
	return cfg
}

// FromContext gets the *Config from the context.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

// ToContextForTest adds Config type values to the context. This should be used for tests only.
func ToContextForTest(ctx context.Context, c *Config) context.Context {
	return toContext(ctx, c)
}

func toContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// DefaultConfigContext attaches a Config with all default values to the
// context.
func DefaultConfigContext(ctx context.Context) context.Context {
	return toContext(ctx, &Config{
		defaults: BuiltinDefaultsConfig(),
	})
}

// Store is based on configmap.UntypedStore and is used to store and watch for
// updates to configuration related to defaults.
// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a configmap.UntypedStore based config store.
//
// logger must be non-nil implementation of configmap.Logger (commonly used
// loggers conform)
//
// onAfterStore is a variadic list of callbacks to run
// after the ConfigMap has been processed and stored.
//
// See also: configmap.NewUntypedStore().
func NewStore(logger configmap.Logger, onAfterStore ...func(name string, value interface{})) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"defaults",
			logger,
			configmap.Constructors{
				DefaultsConfigName: NewDefaultsConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// NewDefaultConfigStore creates a config store populated
// with default keys and values.
func NewDefaultConfigStore(logger configmap.Logger) *Store {
	store := NewStore(logger)
	cfg := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultsConfigName,
		},
	}
	store.OnConfigChanged(cfg)
	return store
}

// ToContext loads the DefaultsConfig from the store into the context.
// If there is an error loading the defaults config, the error is stored on the Config type.
func (s *Store) ToContext(ctx context.Context) context.Context {
	loadedConfig := s.waitToLoad()
	return toContext(ctx, loadedConfig)
}

func (s *Store) waitToLoad() *Config {
	if s.UntypedLoad(DefaultsConfigName) == nil {
		return &Config{
			loadError: errors.New("error loading defaults config, value is nil"),
		}
	}
	return &Config{
		defaults: s.UntypedLoad(DefaultsConfigName).(*DefaultsConfig),
	}
}
