// Copyright 2022 Google LLC
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
	// ComponentName is Name of the component to pull logs from.
	ComponentName string
	// ContainerName is Container name to pull logs from.
	ContainerName string
	// Follow is stream the logs
	Follow bool
	// Labels is Labels to filter the Pods when tailing logs.
	Labels map[string]string
	// NumberLines is number of lines
	NumberLines int
	// Space is the Space to use
	Space string
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

// ComponentName returns the last set value for ComponentName or the empty value
// if not set.
func (opts TailOptions) ComponentName() string {
	return opts.toConfig().ComponentName
}

// ContainerName returns the last set value for ContainerName or the empty value
// if not set.
func (opts TailOptions) ContainerName() string {
	return opts.toConfig().ContainerName
}

// Follow returns the last set value for Follow or the empty value
// if not set.
func (opts TailOptions) Follow() bool {
	return opts.toConfig().Follow
}

// Labels returns the last set value for Labels or the empty value
// if not set.
func (opts TailOptions) Labels() map[string]string {
	return opts.toConfig().Labels
}

// NumberLines returns the last set value for NumberLines or the empty value
// if not set.
func (opts TailOptions) NumberLines() int {
	return opts.toConfig().NumberLines
}

// Space returns the last set value for Space or the empty value
// if not set.
func (opts TailOptions) Space() string {
	return opts.toConfig().Space
}

// Timeout returns the last set value for Timeout or the empty value
// if not set.
func (opts TailOptions) Timeout() time.Duration {
	return opts.toConfig().Timeout
}

// WithTailComponentName creates an Option that sets Name of the component to pull logs from.
func WithTailComponentName(val string) TailOption {
	return func(cfg *tailConfig) {
		cfg.ComponentName = val
	}
}

// WithTailContainerName creates an Option that sets Container name to pull logs from.
func WithTailContainerName(val string) TailOption {
	return func(cfg *tailConfig) {
		cfg.ContainerName = val
	}
}

// WithTailFollow creates an Option that sets stream the logs
func WithTailFollow(val bool) TailOption {
	return func(cfg *tailConfig) {
		cfg.Follow = val
	}
}

// WithTailLabels creates an Option that sets Labels to filter the Pods when tailing logs.
func WithTailLabels(val map[string]string) TailOption {
	return func(cfg *tailConfig) {
		cfg.Labels = val
	}
}

// WithTailNumberLines creates an Option that sets number of lines
func WithTailNumberLines(val int) TailOption {
	return func(cfg *tailConfig) {
		cfg.NumberLines = val
	}
}

// WithTailSpace creates an Option that sets the Space to use
func WithTailSpace(val string) TailOption {
	return func(cfg *tailConfig) {
		cfg.Space = val
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
		WithTailNumberLines(0),
		WithTailSpace("default"),
		WithTailTimeout(time.Second),
	}
}
