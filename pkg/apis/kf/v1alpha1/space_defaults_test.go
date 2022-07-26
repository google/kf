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
	"testing"

	apiconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func defaultContext() context.Context {
	return apiconfig.DefaultConfigContext(context.Background())
}

// sampleConfig overrides ALL values
func sampleConfig() context.Context {
	cfg := apiconfig.CreateConfigForTest(&apiconfig.DefaultsConfig{
		SpaceContainerRegistry: "gcr.io/mycompany/project-xyz",
		SpaceClusterDomains: []apiconfig.DomainTemplate{
			{Domain: "$(SPACE_NAME).custom.example.com"},
		},
	})
	return apiconfig.ToContextForTest(context.Background(), cfg)
}

func TestSpaceSpecNetworkConfig_SetDefaults(t *testing.T) {
	defaultPolicy := SpaceSpecNetworkConfigPolicy{
		Ingress: PermitAllNetworkPolicy,
		Egress:  PermitAllNetworkPolicy,
	}

	cases := testutil.ApisDefaultableTestSuite{
		"builtin": {
			Context: sampleConfig(),
			Input:   &SpaceSpecNetworkConfig{},
			Want: &SpaceSpecNetworkConfig{
				AppNetworkPolicy:   defaultPolicy,
				BuildNetworkPolicy: defaultPolicy,
			},
		},
		"domains retain order": {
			Context: sampleConfig(),
			Input: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{Domain: "z.com"},
					{Domain: "a.com"},
					{Domain: "b.com"},
				},
			},
			Want: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{
						Domain:      "z.com",
						GatewayName: "kf/external-gateway",
					},
					{
						Domain:      "a.com",
						GatewayName: "kf/external-gateway",
					},
					{
						Domain:      "b.com",
						GatewayName: "kf/external-gateway",
					},
				},
				AppNetworkPolicy:   defaultPolicy,
				BuildNetworkPolicy: defaultPolicy,
			},
		},
		"domains retain order while deduplicated": {
			Context: sampleConfig(),
			Input: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{Domain: "example.com"},
					{Domain: "other-example.com"},
					{Domain: "example.com"},
				},
			},
			Want: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{
						Domain:      "example.com",
						GatewayName: "kf/external-gateway",
					},
					{
						Domain:      "other-example.com",
						GatewayName: "kf/external-gateway",
					},
				},
				AppNetworkPolicy:   defaultPolicy,
				BuildNetworkPolicy: defaultPolicy,
			},
		},
		"replace empty domain gateways": {
			Context: sampleConfig(),
			Input: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{Domain: "example.com"},
				},
			},
			Want: &SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{
						Domain:      "example.com",
						GatewayName: "kf/external-gateway",
					},
				},
				AppNetworkPolicy:   defaultPolicy,
				BuildNetworkPolicy: defaultPolicy,
			},
		},
		"don't override set policies": {
			Context: sampleConfig(),
			Input: &SpaceSpecNetworkConfig{
				AppNetworkPolicy: SpaceSpecNetworkConfigPolicy{
					Ingress: "CustomAIPolicy",
					Egress:  "CustomAEPolicy",
				},
				BuildNetworkPolicy: SpaceSpecNetworkConfigPolicy{
					Ingress: "CustomBIPolicy",
					Egress:  "CustomBEPolicy",
				},
			},
			Want: &SpaceSpecNetworkConfig{
				AppNetworkPolicy: SpaceSpecNetworkConfigPolicy{
					Ingress: "CustomAIPolicy",
					Egress:  "CustomAEPolicy",
				},
				BuildNetworkPolicy: SpaceSpecNetworkConfigPolicy{
					Ingress: "CustomBIPolicy",
					Egress:  "CustomBEPolicy",
				},
			},
		},
	}

	cases.Run(t)
}

func TestSpaceSpecBuildConfig_SetDefaults(t *testing.T) {
	cases := testutil.ApisDefaultableTestSuite{
		"builtin": {
			Context: defaultContext(),
			Input:   &SpaceSpecBuildConfig{},
			Want: &SpaceSpecBuildConfig{
				ServiceAccount: DefaultBuildServiceAccountName,
			},
		},
		"custom-config": {
			Context: sampleConfig(),
			Input:   &SpaceSpecBuildConfig{},
			Want: &SpaceSpecBuildConfig{
				ContainerRegistry: "gcr.io/mycompany/project-xyz",
				ServiceAccount:    DefaultBuildServiceAccountName,
			},
		},
		"do not override set values": {
			Context: defaultContext(),
			Input: &SpaceSpecBuildConfig{
				ContainerRegistry: "b",
				ServiceAccount:    "some-other-account",
			},
			Want: &SpaceSpecBuildConfig{
				ContainerRegistry: "b",
				ServiceAccount:    "some-other-account",
			},
		},
	}

	cases.Run(t)
}

func TestSpace_SetDefaults(t *testing.T) {
	// This test just asserts a blank space being defaulted for sanity purposes
	// to ensure nothing new gets added. Each function should also have its own
	// complete tests.
	cases := testutil.ApisDefaultableTestSuite{
		"full-custom-config": {
			Context: sampleConfig(),
			Input: &Space{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-space",
				},
			},
			Want: &Space{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-space",
				},
				Spec: SpaceSpec{
					BuildConfig: SpaceSpecBuildConfig{
						ServiceAccount:    "kf-builder",
						ContainerRegistry: "gcr.io/mycompany/project-xyz",
						Env:               nil,
					},
					NetworkConfig: SpaceSpecNetworkConfig{
						AppNetworkPolicy: SpaceSpecNetworkConfigPolicy{
							Ingress: PermitAllNetworkPolicy,
							Egress:  PermitAllNetworkPolicy,
						},
						BuildNetworkPolicy: SpaceSpecNetworkConfigPolicy{
							Ingress: PermitAllNetworkPolicy,
							Egress:  PermitAllNetworkPolicy,
						},
					},
				},
			},
		},
	}

	cases.Run(t)
}
