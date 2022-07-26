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
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apitesting "knative.dev/pkg/apis/testing"
	"knative.dev/pkg/ptr"
)

func TestSpaceDuckTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    duck.Implementable
	}{
		{
			name: "conditions",
			t:    &duckv1beta1.Conditions{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := duck.VerifyType(&Space{}, test.t)
			if err != nil {
				t.Errorf("VerifyType(Service, %T) = %v", test.t, err)
			}
		})
	}
}

func TestSpaceGeneration(t *testing.T) {
	t.Parallel()
	space := Space{}
	testutil.AssertEqual(t, "empty space generation", int64(0), space.GetGeneration())

	answer := int64(42)
	space.SetGeneration(answer)
	testutil.AssertEqual(t, "GetGeneration", answer, space.GetGeneration())
}

func TestSpaceIsReady(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		status  SpaceStatus
		isReady bool
	}{{
		name:    "empty status should not be ready",
		status:  SpaceStatus{},
		isReady: false,
	}, {
		name: "Different condition type should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: false,
	}, {
		name: "False condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Unknown condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
		isReady: false,
	}, {
		name: "Missing condition status should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type: SpaceConditionReady,
				}},
			},
		},
		isReady: false,
	}, {
		name: "True condition status should be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   SpaceConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status should be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SpaceConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		isReady: true,
	}, {
		name: "Multiple conditions with ready status false should not be ready",
		status: SpaceStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{{
					Type:   "Foo",
					Status: corev1.ConditionTrue,
				}, {
					Type:   SpaceConditionReady,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		isReady: false,
	}}

	for _, tc := range cases {
		testutil.AssertEqual(t, tc.name, tc.isReady, tc.status.IsReady())
	}
}

func initTestStatus(t *testing.T) *SpaceStatus {
	t.Helper()
	status := &SpaceStatus{}
	status.InitializeConditions()

	// sanity check
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionNamespaceReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionBuildServiceAccountReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionBuildSecretReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionRuntimeConfigReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionNetworkConfigReady, t)
	apitesting.CheckConditionOngoing(status.duck(), SpaceConditionBuildConfigReady, t)

	return status
}

func TestPropagateNamespaceStatus_terminating(t *testing.T) {
	t.Parallel()
	status := initTestStatus(t)

	status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating}})

	apitesting.CheckConditionFailed(status.duck(), SpaceConditionReady, t)
	apitesting.CheckConditionFailed(status.duck(), SpaceConditionNamespaceReady, t)
}

func TestSpaceStatus_lifecycle(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		Init func(*SpaceStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *SpaceStatus) {
				status.PropagateIngressGatewayStatus([]corev1.LoadBalancerIngress{{}})
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}})
				status.BuildRoleCondition().MarkSuccess()
				status.BuildRoleBindingCondition().MarkSuccess()
				status.BuildServiceAccountCondition().MarkSuccess()
				status.BuildSecretCondition().MarkSuccess()
				status.NetworkConfigCondition().MarkSuccess()
				status.RuntimeConfigCondition().MarkSuccess()
				status.BuildConfigCondition().MarkSuccess()
				status.AppNetworkPolicyCondition().MarkSuccess()
				status.BuildNetworkPolicyCondition().MarkSuccess()
				status.RoleBindingsCondition().MarkSuccess()
				status.ClusterRoleCondition().MarkSuccess()
				status.ClusterRoleBindingsCondition().MarkSuccess()
				status.IAMPolicyCondition().MarkSuccess()
			},
			ExpectSucceeded: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
				SpaceConditionBuildServiceAccountReady,
				SpaceConditionBuildSecretReady,
				SpaceConditionIngressGatewayReady,
				SpaceConditionNetworkConfigReady,
				SpaceConditionRuntimeConfigReady,
				SpaceConditionBuildConfigReady,
				SpaceConditionAppNetworkPolicyReady,
				SpaceConditionBuildNetworkPolicyReady,
				SpaceConditionRoleBindingsReady,
				SpaceConditionClusterRoleReady,
				SpaceConditionClusterRoleBindingsReady,
				SpaceConditionIAMPolicyReady,
			},
		},
		"terminating namespace": {
			Init: func(status *SpaceStatus) {
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating}})
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionIngressGatewayReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
			},
		},
		"unknown namespace": {
			Init: func(status *SpaceStatus) {
				status.PropagateNamespaceStatus(&corev1.Namespace{Status: corev1.NamespaceStatus{}})
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
				SpaceConditionIngressGatewayReady,
			},
		},
		"ns not owned": {
			Init: func(status *SpaceStatus) {
				status.NamespaceCondition().MarkChildNotOwned("my-ns")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionIngressGatewayReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionNamespaceReady,
			},
		},
		"Build ServiceAccount not owned": {
			Init: func(status *SpaceStatus) {
				status.BuildServiceAccountCondition().MarkChildNotOwned("build-service-account")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
				SpaceConditionIngressGatewayReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionBuildServiceAccountReady,
			},
		},
		"Build Secret not owned": {
			Init: func(status *SpaceStatus) {
				status.BuildSecretCondition().MarkChildNotOwned("build-secret")
			},
			ExpectOngoing: []apis.ConditionType{
				SpaceConditionNamespaceReady,
			},
			ExpectFailed: []apis.ConditionType{
				SpaceConditionReady,
				SpaceConditionBuildSecretReady,
			},
		},
	}

	// XXX: if we start copying state from subresources back to the parent,
	// ensure that the state is updated.

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestStatus(t)

			tc.Init(status)

			for _, exp := range tc.ExpectFailed {
				apitesting.CheckConditionFailed(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectOngoing {
				apitesting.CheckConditionOngoing(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectSucceeded {
				apitesting.CheckConditionSucceeded(status.duck(), exp, t)
			}
		})
	}
}

func TestSpaceStatus_PropagateBuildConfigStatus(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		spaceSpec SpaceSpec
		cfg       *config.Config

		expectStatus v1.ConditionStatus
	}{
		"blank everything": {
			expectStatus: v1.ConditionFalse,
		},
		"removes disabled v2 buildpacks": {
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceBuildpacksV2: config.BuildpackV2List{
					{
						Name: "java_buildpack",
						URL:  "path/to/java",
					},
					{
						Name:     "disabled_buildpack",
						Disabled: true,
						URL:      "path/to/disabled",
					},
				},
			},
			),
			expectStatus: v1.ConditionTrue,
		},
		"unset default": {
			spaceSpec: SpaceSpec{
				BuildConfig: SpaceSpecBuildConfig{
					DefaultToV3Stack: nil,
				},
			},
			cfg:          config.CreateConfigForTest(config.BuiltinDefaultsConfig()),
			expectStatus: v1.ConditionTrue,
		},
		"default override false": {
			spaceSpec: SpaceSpec{
				BuildConfig: SpaceSpecBuildConfig{
					DefaultToV3Stack: ptr.Bool(false),
				},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceDefaultToV3Stack: true,
			}),
			expectStatus: v1.ConditionTrue,
		},
		"default override true": {
			spaceSpec: SpaceSpec{
				BuildConfig: SpaceSpecBuildConfig{
					DefaultToV3Stack: ptr.Bool(true),
				},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceDefaultToV3Stack: false,
			}),
			expectStatus: v1.ConditionTrue,
		},
		"complete flow": {
			spaceSpec: SpaceSpec{
				BuildConfig: SpaceSpecBuildConfig{
					DefaultToV3Stack: ptr.Bool(false),
				},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceBuildpacksV2: config.BuildpackV2List{
					{
						Name: "java_buildpack",
						URL:  "path/to/java",
					},
				},
				SpaceStacksV2: config.StackV2List{
					{
						Name:  "legacy-buildpacks",
						Image: "gcr.io/legacy/buildpacks",
					},
				},
				SpaceStacksV3: config.StackV3List{
					{
						Name:       "google-slim",
						BuildImage: "gcr.io/google/google-slim",
						RunImage:   "gcr.io/google/google-slim",
					},
				},
				SpaceDefaultToV3Stack: true,
			}),
			expectStatus: v1.ConditionTrue,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &SpaceStatus{}
			status.PropagateBuildConfigStatus(tc.spaceSpec, tc.cfg)
			testutil.AssertGoldenJSON(t, "status.buildConfig", status.BuildConfig)
			testutil.AssertNil(t, "condition err", apitesting.CheckCondition(status.duck(), SpaceConditionBuildConfigReady, tc.expectStatus))
		})
	}
}

func TestSpaceStatus_PropagateNetworkConfigStatus(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		specConfig      SpaceSpecNetworkConfig
		cfg             *config.Config
		ingressGateways []corev1.LoadBalancerIngress

		expectStatus v1.ConditionStatus
	}{
		"full test": {
			specConfig: SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{
					{Domain: "$(SPACE_NAME).shell-company.com"},
					{Domain: "$(CLUSTER_INGRESS_IP).prod.shell-company.com"},
				},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceClusterDomains: []config.DomainTemplate{
					{Domain: "$(SPACE_NAME).example.com"},
					{Domain: "$(CLUSTER_INGRESS_IP).example.com"},
					{Domain: "$(NO_REPLACE).example.com"},
					{Domain: "example.com"},
					{Domain: "$(SPACE_NAME)-apps.internal", GatewayName: "kf/internal-gateway"},
				},
			}),
			ingressGateways: []corev1.LoadBalancerIngress{
				{IP: "192.168.0.1"},
			},
			expectStatus: v1.ConditionTrue,
		},
		"blank config": {
			specConfig: SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceClusterDomains: []config.DomainTemplate{
					{Domain: "$(SPACE_NAME).$(CLUSTER_INGRESS_IP).example.com"},
					{Domain: "example.com"},
				},
			}),
			ingressGateways: []corev1.LoadBalancerIngress{
				{IP: "192.168.0.1"},
			},
			expectStatus: v1.ConditionTrue,
		},
		"no ingress IP": {
			specConfig: SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{},
			},
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				SpaceClusterDomains: []config.DomainTemplate{
					{Domain: "$(SPACE_NAME).$(CLUSTER_INGRESS_IP).example.com"},
					{Domain: "example.com"},
				},
			}),
			ingressGateways: []corev1.LoadBalancerIngress{},
			expectStatus:    v1.ConditionTrue,
		},
		"config not loaded": {
			specConfig: SpaceSpecNetworkConfig{
				Domains: []SpaceDomain{},
			},
			cfg:             nil,
			ingressGateways: nil,
			expectStatus:    v1.ConditionFalse,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &SpaceStatus{}
			status.IngressGateways = tc.ingressGateways
			status.PropagateNetworkConfigStatus(tc.specConfig, tc.cfg, "test-space")
			var err error
			configDefaults := &config.DefaultsConfig{}
			if tc.cfg == nil {
				configDefaults = nil
			} else {
				configDefaults, err = tc.cfg.Defaults()
				testutil.AssertNil(t, "err", err)
			}
			testutil.AssertGoldenJSONContext(t, "status.NetworkConfig", status.NetworkConfig, map[string]interface{}{
				"config.defaults":              configDefaults,
				"space.spec.networkConfig":     tc.specConfig,
				"space.status.ingressGateways": tc.ingressGateways,
			})
			testutil.AssertNil(t, "condition err", apitesting.CheckCondition(status.duck(), SpaceConditionNetworkConfigReady, tc.expectStatus))
		})
	}
}
