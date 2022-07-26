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

package genericcli

import (
	"context"

	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

// Type is the interface components need to satisfy to generate generic CLI
// utilities for them.
type Type interface {
	Namespaced() bool
	GroupVersionResource(context.Context) schema.GroupVersionResource
	GroupVersionKind(context.Context) schema.GroupVersionKind
	FriendlyName() string
}

// GetResourceInterface is a helper method to return the Resource interface of a given client.
func GetResourceInterface(ctx context.Context, t Type, client dynamic.Interface, ns string) dynamic.ResourceInterface {
	if t.Namespaced() {
		return client.Resource(t.GroupVersionResource(ctx)).Namespace(ns)
	}

	return client.Resource(t.GroupVersionResource(ctx))
}

// KubernetesType is an implementation of Type that relies on hard-coded
// values.
type KubernetesType struct {
	NsScoped bool
	Group    string
	Version  string
	Kind     string
	Resource string
	KfName   string
}

var _ Type = (*KubernetesType)(nil)

// Namespaced implements Type.Namespaced
func (g *KubernetesType) Namespaced() bool {
	return g.NsScoped
}

// GroupVersionResource implements Type.GroupVersionResource
func (g *KubernetesType) GroupVersionResource(context.Context) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.Group,
		Version:  g.Version,
		Resource: g.Resource,
	}
}

// GroupVersionKind implements Type.GroupVersionKind
func (g *KubernetesType) GroupVersionKind(context.Context) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.Group,
		Version: g.Version,
		Kind:    g.Kind,
	}
}

// FriendlyName implements Type.FriendlyName
func (g *KubernetesType) FriendlyName() string {
	return g.KfName
}

// ValidArgsFunction creates a cobra.ValidArgsFunction for the generic Type t.
func ValidArgsFunction(t Type, p *config.KfParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()
		return completion.GenericCompletionFn(
			GetResourceInterface(ctx, t, dynamicclient.Get(cmd.Context()), p.Space),
		)(cmd, args, toComplete)
	}
}
