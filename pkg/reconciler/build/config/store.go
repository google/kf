// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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

type Config struct {
	secrets   *SecretsConfig
	loadError error
}

// Secrets returns the SecretsConfig and an error if there was an issue loading the config.
func (c *Config) Secrets() (*SecretsConfig, error) {
	return c.secrets, c.loadError
}

// FromContext gets the *Config from the context.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

func toContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// Store is based on configmap.UntypedStore and is used to store and watch for
// updates to configuration related to Secrets.
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
func NewStore(
	logger configmap.Logger,
	onAfterStore ...func(name string, value interface{}),
) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"secrets",
			logger,
			configmap.Constructors{
				SecretsConfigName: NewSecretsConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// NewSecretsConfigStore creates a secrets config store populated
// with default keys and values.
func NewSecretsConfigStore(logger configmap.Logger) *Store {
	store := NewStore(logger)
	secretsConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: SecretsConfigName,
		},
	}
	store.OnConfigChanged(secretsConfig)
	return store
}

// ToContext loads the SecretsConfig from the store into the context.
// If there is an error loading the secrets config, the error is stored on the Config type.
func (s *Store) ToContext(ctx context.Context) context.Context {
	loadedConfig := s.waitToLoad()
	return toContext(ctx, loadedConfig)
}

func (s *Store) waitToLoad() *Config {
	if s.UntypedLoad(SecretsConfigName) == nil {
		return &Config{
			loadError: errors.New("error loading secrets config, value is nil"),
		}
	}
	return &Config{
		secrets: s.UntypedLoad(SecretsConfigName).(*SecretsConfig),
	}
}
