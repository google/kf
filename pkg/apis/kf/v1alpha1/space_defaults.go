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

package v1alpha1

import (
	"context"

	apiconfig "github.com/google/kf/v2/pkg/apis/kf/config"
)

const (
	// DefaultBuildServiceAccountName is the name of the service-account created
	// by spaces with the intent of being used to hold connection credentials to
	// the necessary back-end systems.
	DefaultBuildServiceAccountName = "kf-builder"

	// KfExternalIngressGateway holds the gateway for Kf's external HTTP ingress.
	KfExternalIngressGateway = "kf/external-gateway"
)

// SetDefaults implements apis.Defaultable
func (k *Space) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpec) SetDefaults(ctx context.Context) {
	k.BuildConfig.SetDefaults(ctx)
	k.RuntimeConfig.SetDefaults(ctx)
	k.NetworkConfig.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecBuildConfig) SetDefaults(ctx context.Context) {
	if k.ServiceAccount == "" {
		k.ServiceAccount = DefaultBuildServiceAccountName
	}

	if k.ContainerRegistry == "" {
		configDefaults, err := apiconfig.FromContext(ctx).Defaults()
		if err == nil {
			k.ContainerRegistry = configDefaults.SpaceContainerRegistry
		}
	}
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecRuntimeConfig) SetDefaults(ctx context.Context) {
	// no defaults to set
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecNetworkConfig) SetDefaults(ctx context.Context) {
	k.Domains = StableDeduplicateSpaceDomainList(k.Domains)
	k.DefaultSpaceDomainGateways(ctx)
	k.AppNetworkPolicy.SetDefaults(ctx)
	k.BuildNetworkPolicy.SetDefaults(ctx)
}

// DefaultSpaceDomainGateways replaces missing gatewayNames with the kf
// default external gateway.
func (k *SpaceSpecNetworkConfig) DefaultSpaceDomainGateways(ctx context.Context) {
	for i := range k.Domains {
		if k.Domains[i].GatewayName == "" {
			k.Domains[i].GatewayName = KfExternalIngressGateway
		}
	}
}

// SetDefaults implements apis.Defaultable
func (s *SpaceSpecNetworkConfigPolicy) SetDefaults(ctx context.Context) {
	if s.Ingress == "" {
		s.Ingress = PermitAllNetworkPolicy
	}

	if s.Egress == "" {
		s.Egress = PermitAllNetworkPolicy
	}
}
