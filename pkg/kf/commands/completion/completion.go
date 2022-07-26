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

package completion

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

const (
	// AppCompletion is the type for completing apps
	AppCompletion = "apps"

	// BuildCompletion is the type for completing builds
	BuildCompletion = "builds"

	// SpaceCompletion is the type for completing spaces
	SpaceCompletion = "spaces"

	// NetworkPolicyCompletion is the type for completing networkpolicies
	NetworkPolicyCompletion = "networkpolicies"
)

var namespacedTypes = map[string]schema.GroupVersionResource{
	AppCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "apps",
	},

	BuildCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "builds",
	},

	NetworkPolicyCompletion: {
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "networkpolicies",
	},
}

var globalTypes = map[string]schema.GroupVersionResource{
	SpaceCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "spaces",
	},
}

// AppCompletionFn is a cobra.ValidArgsFunction completer for Apps.
func AppCompletionFn(p *config.KfParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return k8sTypeCompletionFn(p, AppCompletion)
}

// BuildCompletionFn is a cobra.ValidArgsFunction completer for Builds.
func BuildCompletionFn(p *config.KfParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return k8sTypeCompletionFn(p, BuildCompletion)
}

// SpaceCompletionFn is a cobra.ValidArgsFunction completer for Spaces.
func SpaceCompletionFn(p *config.KfParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return k8sTypeCompletionFn(p, SpaceCompletion)
}

// NetworkPoliycCompletionFn is a cobra.ValidArgsFunction completer for NetworkPolicies.
func NetworkPolicyCompletionFn(p *config.KfParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return k8sTypeCompletionFn(p, NetworkPolicyCompletion)
}

// GenericCompletionFn is a cobra.ValidArgsFunction completer for generic cli
// commands. The resource type is determined by the provided ResourceInterface.
func GenericCompletionFn(client dynamic.ResourceInterface) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionFn(client)
}

// completionFn is a cobra.ValidArgsFunction completer factory for a given
// ResourceInterface. It suggests any resource of the given type that begins
// with the string typed by the user so far in sorted order.
//
// If any error is encountered it does not provide any suggestions.
func completionFn(client dynamic.ResourceInterface) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		resources, err := client.List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var names []string
		for _, resource := range resources.Items {
			name := resource.GetName()
			if strings.HasPrefix(name, toComplete) {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		return names, cobra.ShellCompDirectiveNoFileComp
	}
}

// k8sTypeCompletionFn is a helper to generate a completer function for the
// provided k8sType. It handles creating the associated ResourceInterface
// client.
func k8sTypeCompletionFn(p *config.KfParams, k8sType string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := getResourceInterface(cmd.Context(), p, k8sType)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return completionFn(client)(cmd, args, toComplete)
	}
}

func getResourceInterface(ctx context.Context, p *config.KfParams, k8sType string) (dynamic.ResourceInterface, error) {
	client := dynamicclient.Get(ctx)
	if resource, ok := namespacedTypes[k8sType]; ok {
		if err := p.ValidateSpaceTargeted(); err != nil {
			return nil, err
		}
		return client.Resource(resource).Namespace(p.Space), nil
	}

	if resource, ok := globalTypes[k8sType]; ok {
		return client.Resource(resource), nil
	}

	return nil, fmt.Errorf("unknown type: %s", k8sType)
}
