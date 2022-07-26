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
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestSpaceValidation(t *testing.T) {
	goodBuildConfig := SpaceSpecBuildConfig{
		ContainerRegistry: "gcr.io/test",
		ServiceAccount:    DefaultBuildServiceAccountName,
	}

	goodNetworkPolicy := SpaceSpecNetworkConfigPolicy{
		Ingress: DenyAllNetworkPolicy,
		Egress:  DenyAllNetworkPolicy,
	}

	goodNetworkConfig := SpaceSpecNetworkConfig{
		Domains:            []SpaceDomain{{Domain: "example.com", GatewayName: "kf/some-gateway"}},
		AppNetworkPolicy:   goodNetworkPolicy,
		BuildNetworkPolicy: goodNetworkPolicy,
	}

	goodSpaceSpec := SpaceSpec{
		BuildConfig:   goodBuildConfig,
		NetworkConfig: goodNetworkConfig,
	}
	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	cases := map[string]struct {
		space *Space
		want  *apis.FieldError
	}{
		"good": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec:       goodSpaceSpec,
			},
		},
		"invalid ObjectMeta": {
			space: &Space{
				ObjectMeta: badMeta,
				Spec:       goodSpaceSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"reserved name: kf": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "kf"},
				Spec:       goodSpaceSpec,
			},
			want: apis.ErrInvalidValue("kf", "name"),
		},
		"reserved name: default": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
				Spec:       goodSpaceSpec,
			},
			want: apis.ErrInvalidValue("default", "name"),
		},
		"no container registry": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: goodNetworkConfig,
					BuildConfig: SpaceSpecBuildConfig{
						ServiceAccount: DefaultBuildServiceAccountName,
					},
				},
			},
			want: apis.ErrMissingField("spec.buildConfig.containerRegistry"),
		},
		"no service account": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: goodNetworkConfig,
					BuildConfig: SpaceSpecBuildConfig{
						ContainerRegistry: "gcr.io/test",
					},
				},
			},
			want: apis.ErrMissingField("spec.buildConfig.serviceAccount"),
		},
		"no domains": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					BuildConfig: goodBuildConfig,
					NetworkConfig: SpaceSpecNetworkConfig{
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
				},
			},
			// no domains is okay
		},
		"no duplicated domains": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "kf/some-gateway",
							},
							{
								Domain:      "example2.com",
								GatewayName: "kf/some-gateway",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			// should be okay
		},
		"duplicate domains": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "kf/some-gateway",
							},
							{
								Domain:      "google.com",
								GatewayName: "kf/some-gateway",
							},
							{
								Domain:      "example.com",
								GatewayName: "kf/some-gateway",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: errDuplicateValue("example.com", "spec.networkConfig.domains[2].domain"),
		},
		"bad network policy": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					BuildConfig: goodBuildConfig,
					NetworkConfig: SpaceSpecNetworkConfig{
						AppNetworkPolicy: SpaceSpecNetworkConfigPolicy{
							Ingress: "badappingress",
							Egress:  "badappegress",
						},
						BuildNetworkPolicy: SpaceSpecNetworkConfigPolicy{
							Ingress: "badbldingress",
							Egress:  "badbldegress",
						},
					},
				},
			},
			want: (*apis.FieldError)(nil).Also(
				ErrInvalidEnumValue("badappegress", "spec.networkConfig.appNetworkPolicy.egress", []string{DenyAllNetworkPolicy, PermitAllNetworkPolicy}),
				ErrInvalidEnumValue("badappingress", "spec.networkConfig.appNetworkPolicy.ingress", []string{DenyAllNetworkPolicy, PermitAllNetworkPolicy}),
				ErrInvalidEnumValue("badbldegress", "spec.networkConfig.buildNetworkPolicy.egress", []string{DenyAllNetworkPolicy, PermitAllNetworkPolicy}),
				ErrInvalidEnumValue("badbldingress", "spec.networkConfig.buildNetworkPolicy.ingress", []string{DenyAllNetworkPolicy, PermitAllNetworkPolicy}),
			),
		},
		"custom gateways": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "kf/some-gateway",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
		},
		"custom gateways missing gatewayName": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: (*apis.FieldError)(nil).Also(
				apis.ErrMissingField("gatewayName").ViaFieldIndex("spec.networkConfig.domains", 0),
			),
		},
		"custom gateways missing namespace": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "some-gateway-namespace-assumed-kf",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: (*apis.FieldError)(nil).Also(
				(&apis.FieldError{
					Message: "Invalid gatewayName",
					Paths:   []string{"gatewayName"},
					Details: "Namespace prefix is missing",
				}).ViaFieldIndex("spec.networkConfig.domains", 0),
			),
		},
		"custom gateways invalid namespace prefix": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "/broken",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: (*apis.FieldError)(nil).Also(
				(&apis.FieldError{
					Message: "Invalid gatewayName",
					Paths:   []string{"gatewayName"},
					Details: "Gateway Namespace was missing",
				}).ViaFieldIndex("spec.networkConfig.domains", 0),
			),
		},
		"custom gateways invalid name suffix": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "kf/",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: (*apis.FieldError)(nil).Also(
				(&apis.FieldError{
					Message: "Invalid gatewayName",
					Paths:   []string{"gatewayName"},
					Details: "Gateway name was missing",
				}).ViaFieldIndex("spec.networkConfig.domains", 0),
			),
		},
		"custom gateways invalid namespace not kf": {
			space: &Space{
				ObjectMeta: metav1.ObjectMeta{Name: "valid"},
				Spec: SpaceSpec{
					NetworkConfig: SpaceSpecNetworkConfig{
						Domains: []SpaceDomain{
							{
								Domain:      "example.com",
								GatewayName: "not-kf/gateway",
							},
						},
						AppNetworkPolicy:   goodNetworkPolicy,
						BuildNetworkPolicy: goodNetworkPolicy,
					},
					BuildConfig: goodBuildConfig,
				},
			},
			want: (*apis.FieldError)(nil).Also(
				(&apis.FieldError{
					Message: "Invalid namespace for gatewayName",
					Paths:   []string{"gatewayName"},
					Details: "Only the kf namespace is allowed",
				}).ViaFieldIndex("spec.networkConfig.domains", 0),
			),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.space.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
