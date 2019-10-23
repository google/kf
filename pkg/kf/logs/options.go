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

// This file was generated with option-builder.go, DO NOT EDIT IT.

package logs

import (
	time "time"
)

type tailConfig struct {
	// Follow is stream the logs
	Follow bool
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// NumberLines is number of lines
	NumberLines int
	// Timeout is How much time to wait before giving up when not following.
	Timeout time.Duration
}

// TailOption is a single option for configuring a tailConfig
type TailOption func(*tailConfig)

// TailOptions is a configuration set defining a tailConfig
type TailOptions []TailOption

// toConfig applies all the options to a new tailConfig and returns it.
func (opts TailOptions) toConfig() tailConfig {
	cfg := tailConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new TailOptions with the contents of other overriding
// the values set in this TailOptions.
func (opts TailOptions) Extend(other TailOptions) TailOptions {
	var out TailOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Follow returns the last set value for Follow or the empty value
// if not set.
func (opts TailOptions) Follow() bool {
	return opts.toConfig().Follow
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts TailOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// NumberLines returns the last set value for NumberLines or the empty value
// if not set.
func (opts TailOptions) NumberLines() int {
	return opts.toConfig().NumberLines
}

// Timeout returns the last set value for Timeout or the empty value
// if not set.
func (opts TailOptions) Timeout() time.Duration {
	return opts.toConfig().Timeout
}

// WithTailFollow creates an Option that sets stream the logs
func WithTailFollow(val bool) TailOption {
	return func(cfg *tailConfig) {
		cfg.Follow = val
	}
}

// WithTailNamespace creates an Option that sets the Kubernetes namespace to use
func WithTailNamespace(val string) TailOption {
	return func(cfg *tailConfig) {
		cfg.Namespace = val
	}
}

// WithTailNumberLines creates an Option that sets number of lines
func WithTailNumberLines(val int) TailOption {
	return func(cfg *tailConfig) {
		cfg.NumberLines = val
	}
}

// WithTailTimeout creates an Option that sets How much time to wait before giving up when not following.
func WithTailTimeout(val time.Duration) TailOption {
	return func(cfg *tailConfig) {
		cfg.Timeout = val
	}
}

// TailOptionDefaults gets the default values for Tail.
func TailOptionDefaults() TailOptions {
	return TailOptions{
		WithTailNamespace("default"),
		WithTailNumberLines(10),
		WithTailTimeout(time.Second),
	}
}
