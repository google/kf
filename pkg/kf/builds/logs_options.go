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

package builds

import (
	"context"
	"io"
	"os"
)

type logsConfig struct {
	// Context is the parent context for the log watcher goroutines
	Context context.Context
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Output is the io.Writer to write the logs to
	Output io.Writer
}

// LogsOption is a single option for configuring a logsConfig
type LogsOption func(*logsConfig)

// LogsOptions is a configuration set defining a logsConfig
type LogsOptions []LogsOption

// toConfig applies all the options to a new logsConfig and returns it.
func (opts LogsOptions) toConfig() logsConfig {
	cfg := logsConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new LogsOptions with the contents of other overriding
// the values set in this LogsOptions.
func (opts LogsOptions) Extend(other LogsOptions) LogsOptions {
	var out LogsOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Context returns the last set value for Context or the empty value
// if not set.
func (opts LogsOptions) Context() context.Context {
	return opts.toConfig().Context
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts LogsOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Output returns the last set value for Output or the empty value
// if not set.
func (opts LogsOptions) Output() io.Writer {
	return opts.toConfig().Output
}

// WithLogsContext creates an Option that sets the parent context for the log watcher goroutines
func WithLogsContext(val context.Context) LogsOption {
	return func(cfg *logsConfig) {
		cfg.Context = val
	}
}

// WithLogsNamespace creates an Option that sets the Kubernetes namespace to use
func WithLogsNamespace(val string) LogsOption {
	return func(cfg *logsConfig) {
		cfg.Namespace = val
	}
}

// WithLogsOutput creates an Option that sets the io.Writer to write the logs to
func WithLogsOutput(val io.Writer) LogsOption {
	return func(cfg *logsConfig) {
		cfg.Output = val
	}
}

// LogsOptionDefaults gets the default values for Logs.
func LogsOptionDefaults() LogsOptions {
	return LogsOptions{
		WithLogsContext(context.Background()),
		WithLogsNamespace("default"),
		WithLogsOutput(os.Stdout),
	}
}
