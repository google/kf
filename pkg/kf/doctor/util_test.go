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

package doctor

import (
	"context"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

func TestDiscoverControllerNamespaces(t *testing.T) {
	clientConfig := func(name, ns string) admissionv1.WebhookClientConfig {
		return admissionv1.WebhookClientConfig{
			Service: &admissionv1.ServiceReference{
				Name:      name,
				Namespace: ns,
			},
		}
	}

	validatingConfig := func(name string, configs ...admissionv1.WebhookClientConfig) *admissionv1.ValidatingWebhookConfiguration {
		config := &admissionv1.ValidatingWebhookConfiguration{}
		config.Name = name

		for _, cfg := range configs {
			config.Webhooks = append(config.Webhooks, admissionv1.ValidatingWebhook{
				ClientConfig: cfg,
			})
		}

		return config
	}

	mutatingConfig := func(name string, configs ...admissionv1.WebhookClientConfig) *admissionv1.MutatingWebhookConfiguration {
		config := &admissionv1.MutatingWebhookConfiguration{}
		config.Name = name

		for _, cfg := range configs {
			config.Webhooks = append(config.Webhooks, admissionv1.MutatingWebhook{
				ClientConfig: cfg,
			})
		}

		return config
	}

	cases := map[string]struct {
		objects        []runtime.Object
		wantNamespaces []string
	}{
		"none": {
			objects:        nil,
			wantNamespaces: []string{},
		},
		"validating": {
			objects: []runtime.Object{
				validatingConfig("istio", clientConfig("validator", "istio-system"), clientConfig("validator2", "istio-system")),
				validatingConfig("kf", clientConfig("webhook", "kf"), clientConfig("webhook", "some-k8s-service")),
			},
			wantNamespaces: []string{"istio-system", "kf", "some-k8s-service"},
		},
		"mutating": {
			objects: []runtime.Object{
				mutatingConfig("istio", clientConfig("validator", "istio-system"), clientConfig("validator2", "istio-system")),
				mutatingConfig("kf", clientConfig("webhook", "kf"), clientConfig("webhook", "some-k8s-service")),
			},
			wantNamespaces: []string{"istio-system", "kf", "some-k8s-service"},
		},

		"both": {
			objects: []runtime.Object{
				validatingConfig("validator", clientConfig("validator", "validating-ns")),
				mutatingConfig("mutator", clientConfig("mutator", "mutating-ns")),
			},
			wantNamespaces: []string{"mutating-ns", "validating-ns"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			client := fakek8s.NewSimpleClientset(tc.objects...)

			gotNamespaces := DiscoverControllerNamespaces(context.Background(), client)

			testutil.AssertEqual(t, "namespaces", tc.wantNamespaces, gotNamespaces)
		})
	}
}
