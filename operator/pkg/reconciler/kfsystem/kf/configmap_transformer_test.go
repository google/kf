// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kf

import (
	"context"
	"kf-operator/pkg/apis/kfsystem/kf"
	"kf-operator/pkg/apis/kfsystem/v1alpha1"
	"testing"

	"kf-operator/pkg/testing/k8s"
	mftesting "kf-operator/pkg/testing/manifestival"

	kfconfig "kf-operator/pkg/reconciler/kfsystem/kf/config"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/go-cmp/cmp"
)

func TestConfigSecretsTransformer(t *testing.T) {
	tests := []struct {
		name   string
		in     mftesting.Object
		config v1alpha1.SecretSpec
		want   *unstructured.Unstructured
	}{{
		name: "not a configmap",
		in:   k8s.Deployment("test-deployment"),
		want: mftesting.ToUnstructured(k8s.Deployment("test-deployment")),
	}, {
		name: "not config secret",
		in:   k8s.ConfigMap("config-defaults"),
		want: mftesting.ToUnstructured(k8s.ConfigMap("config-defaults")),
	}, {
		name: "config secret, with workload identity data",
		in:   k8s.ConfigMap(kfconfig.SecretsConfigName),
		config: v1alpha1.SecretSpec{
			WorkloadIdentity: &v1alpha1.SecretWorkloadIdentity{
				GoogleServiceAccount: "testAccount",
				GoogleProjectID:      "testProjectId",
			}},
		want: mftesting.ToUnstructured(
			k8s.ConfigMap(
				kfconfig.SecretsConfigName,
				k8s.WithData(kfconfig.GoogleServiceAccountKey, getGoogleServiceAccount("testAccount", "testProjectId")),
				k8s.WithData(kfconfig.GoogleProjectIDKey, "testProjectId"),
			)),
	}, {
		name: "config secret, with build secret",
		in:   k8s.ConfigMap(kfconfig.SecretsConfigName),
		config: v1alpha1.SecretSpec{
			Build: &v1alpha1.SecretBuild{
				ImagePushSecretName: "testSecretName",
			}},
		want: mftesting.ToUnstructured(
			k8s.ConfigMap(
				kfconfig.SecretsConfigName,
				k8s.WithData(kfconfig.BuildImagePushSecretKey, "testSecretName"),
			)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.in)

			transformer := ConfigSecretsTransform(
				context.Background(),
				tt.config,
			)

			err := transformer(u)
			if err != nil {
				t.Error("Got error", err)
			}

			if diff := cmp.Diff(tt.want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}

func TestConfigDefaultsTransformer(t *testing.T) {
	tests := []struct {
		name   string
		in     mftesting.Object
		config kf.DefaultsConfig
		want   *unstructured.Unstructured
	}{{
		name: "not a configmap",
		in:   k8s.Deployment("test-deployment"),
		want: mftesting.ToUnstructured(k8s.Deployment("test-deployment")),
	}, {
		name: "not config defaults",
		in:   k8s.ConfigMap("config-something"),
		want: mftesting.ToUnstructured(k8s.ConfigMap("config-something")),
	}, {
		name: "config defaults, with container registry",
		in:   k8s.ConfigMap(kf.DefaultsConfigName, k8s.WithData("buildGolangImage", "buildImage"), k8s.WithData("extraField", "test")),
		config: kf.DefaultsConfig{
			SpaceContainerRegistry: "test-registry",
		},
		want: mftesting.ToUnstructured(
			k8s.ConfigMap(
				kf.DefaultsConfigName,
				k8s.WithData("buildGolangImage", "buildImage"),
				k8s.WithData("spaceContainerRegistry", "test-registry"),
				k8s.WithData("spaceDefaultToV3Stack", "false\n"),
				k8s.WithData("extraField", "test"),
				k8s.WithData("buildDisableIstioSidecar", "false\n"),
			)),
	}, {
		name:   "config defaults, existing configs not overriden",
		in:     k8s.ConfigMap(kf.DefaultsConfigName, k8s.WithData("spaceClusterDomains", "- domain: testDomain\n  gatewayName: gateway\n")),
		config: kf.DefaultsConfig{},
		want: mftesting.ToUnstructured(
			k8s.ConfigMap(
				kf.DefaultsConfigName,
				k8s.WithData("spaceClusterDomains", "- domain: testDomain\n  gatewayName: gateway\n"),
				k8s.WithData("spaceDefaultToV3Stack", "false\n"),
				k8s.WithData("buildDisableIstioSidecar", "false\n"),
			)),
	}, {
		name: "config defaults, with featureflags",
		in:   k8s.ConfigMap(kf.DefaultsConfigName, k8s.WithData("extraField", "test")),
		config: kf.DefaultsConfig{
			SpaceClusterDomains: []kf.DomainTemplate{
				{
					Domain:      "testDomain",
					GatewayName: "gateway",
				},
			},
		},
		want: mftesting.ToUnstructured(
			k8s.ConfigMap(
				kf.DefaultsConfigName,
				k8s.WithData("spaceClusterDomains", "- domain: testDomain\n  gatewayName: gateway\n"),
				k8s.WithData("spaceDefaultToV3Stack", "false\n"),
				k8s.WithData("extraField", "test"),
				k8s.WithData("buildDisableIstioSidecar", "false\n"),
			)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.in)

			transformer := ConfigDefaultsTransform(
				context.Background(),
				tt.config,
			)

			err := transformer(u)
			if err != nil {
				t.Error("Got error", err)
			}

			if diff := cmp.Diff(tt.want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}
