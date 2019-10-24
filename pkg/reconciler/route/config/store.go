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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/configmap"
)

type cfgKey struct{}

// +k8s:deepcopy-gen=false
type Config struct {
	Routing *RoutingConfig
}

func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// Store is based on configmap.UntypedStore and is used to store and watch for
// updates to configuration related to routes.
//
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
			"routing",
			logger,
			configmap.Constructors{
				RoutingConfigName: NewRoutingConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

func NewDefaultConfigStore(logger configmap.Logger) *Store {
	store := NewStore(logger)
	routingConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: RoutingConfigName,
		},
	}
	store.OnConfigChanged(routingConfig)
	return store
}

func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

func (s *Store) Load() *Config {
	return &Config{
		Routing: s.UntypedLoad(RoutingConfigName).(*RoutingConfig),
	}
}
