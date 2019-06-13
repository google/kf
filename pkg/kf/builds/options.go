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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

type createConfig struct {
	// Args is the arguments to the build template
	Args map[string]string
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Owner is a reference to the owner of this build
	Owner *v1.OwnerReference
	// ServiceAccount is the service account to run as
	ServiceAccount string
	// SourceImage is a Kontext source image to seed this build with
	SourceImage string
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

// Args returns the last set value for Args or the empty value
// if not set.
func (opts CreateOptions) Args() map[string]string {
	return opts.toConfig().Args
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts CreateOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Owner returns the last set value for Owner or the empty value
// if not set.
func (opts CreateOptions) Owner() *v1.OwnerReference {
	return opts.toConfig().Owner
}

// ServiceAccount returns the last set value for ServiceAccount or the empty value
// if not set.
func (opts CreateOptions) ServiceAccount() string {
	return opts.toConfig().ServiceAccount
}

// SourceImage returns the last set value for SourceImage or the empty value
// if not set.
func (opts CreateOptions) SourceImage() string {
	return opts.toConfig().SourceImage
}

// WithCreateArgs creates an Option that sets the arguments to the build template
func WithCreateArgs(val map[string]string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Args = val
	}
}

// WithCreateNamespace creates an Option that sets the Kubernetes namespace to use
func WithCreateNamespace(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.Namespace = val
	}
}

// WithCreateOwner creates an Option that sets a reference to the owner of this build
func WithCreateOwner(val *v1.OwnerReference) CreateOption {
	return func(cfg *createConfig) {
		cfg.Owner = val
	}
}

// WithCreateServiceAccount creates an Option that sets the service account to run as
func WithCreateServiceAccount(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.ServiceAccount = val
	}
}

// WithCreateSourceImage creates an Option that sets a Kontext source image to seed this build with
func WithCreateSourceImage(val string) CreateOption {
	return func(cfg *createConfig) {
		cfg.SourceImage = val
	}
}

// CreateOptionDefaults gets the default values for Create.
func CreateOptionDefaults() CreateOptions {
	return CreateOptions{
		WithCreateNamespace("default"),
	}
}

type statusConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// StatusOption is a single option for configuring a statusConfig
type StatusOption func(*statusConfig)

// StatusOptions is a configuration set defining a statusConfig
type StatusOptions []StatusOption

// toConfig applies all the options to a new statusConfig and returns it.
func (opts StatusOptions) toConfig() statusConfig {
	cfg := statusConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new StatusOptions with the contents of other overriding
// the values set in this StatusOptions.
func (opts StatusOptions) Extend(other StatusOptions) StatusOptions {
	var out StatusOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts StatusOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithStatusNamespace creates an Option that sets the Kubernetes namespace to use
func WithStatusNamespace(val string) StatusOption {
	return func(cfg *statusConfig) {
		cfg.Namespace = val
	}
}

// StatusOptionDefaults gets the default values for Status.
func StatusOptionDefaults() StatusOptions {
	return StatusOptions{
		WithStatusNamespace("default"),
	}
}

type deleteConfig struct {
	// Namespace is the Kubernetes namespace to use
	Namespace string
}

// DeleteOption is a single option for configuring a deleteConfig
type DeleteOption func(*deleteConfig)

// DeleteOptions is a configuration set defining a deleteConfig
type DeleteOptions []DeleteOption

// toConfig applies all the options to a new deleteConfig and returns it.
func (opts DeleteOptions) toConfig() deleteConfig {
	cfg := deleteConfig{}

	for _, v := range opts {
		v(&cfg)
	}

	return cfg
}

// Extend creates a new DeleteOptions with the contents of other overriding
// the values set in this DeleteOptions.
func (opts DeleteOptions) Extend(other DeleteOptions) DeleteOptions {
	var out DeleteOptions
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts DeleteOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// WithDeleteNamespace creates an Option that sets the Kubernetes namespace to use
func WithDeleteNamespace(val string) DeleteOption {
	return func(cfg *deleteConfig) {
		cfg.Namespace = val
	}
}

// DeleteOptionDefaults gets the default values for Delete.
func DeleteOptionDefaults() DeleteOptions {
	return DeleteOptions{
		WithDeleteNamespace("default"),
	}
}

type tailConfig struct {
	// Context is
	Context context.Context
	// Namespace is the Kubernetes namespace to use
	Namespace string
	// Writer is
	Writer io.Writer
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

// Context returns the last set value for Context or the empty value
// if not set.
func (opts TailOptions) Context() context.Context {
	return opts.toConfig().Context
}

// Namespace returns the last set value for Namespace or the empty value
// if not set.
func (opts TailOptions) Namespace() string {
	return opts.toConfig().Namespace
}

// Writer returns the last set value for Writer or the empty value
// if not set.
func (opts TailOptions) Writer() io.Writer {
	return opts.toConfig().Writer
}

// WithTailContext creates an Option that sets
func WithTailContext(val context.Context) TailOption {
	return func(cfg *tailConfig) {
		cfg.Context = val
	}
}

// WithTailNamespace creates an Option that sets the Kubernetes namespace to use
func WithTailNamespace(val string) TailOption {
	return func(cfg *tailConfig) {
		cfg.Namespace = val
	}
}

// WithTailWriter creates an Option that sets
func WithTailWriter(val io.Writer) TailOption {
	return func(cfg *tailConfig) {
		cfg.Writer = val
	}
}

// TailOptionDefaults gets the default values for Tail.
func TailOptionDefaults() TailOptions {
	return TailOptions{
		WithTailContext(context.Background()),
		WithTailNamespace("default"),
		WithTailWriter(os.Stdout),
	}
}
