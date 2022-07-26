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

package k8s

import (
	mfTesting "kf-operator/pkg/testing/manifestival"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// ValidatingWebhookConfigurationOption enables further configuration of a ValidatingWebhookConfiguration.
type ValidatingWebhookConfigurationOption func(*admissionregistrationv1.ValidatingWebhookConfiguration)

// ManifestivalValidatingWebhookConfiguration creates a ValidatingWebhookConfigurationOption owned by manifestival.
func ManifestivalValidatingWebhookConfiguration(name string, vwcho ...ValidatingWebhookConfigurationOption) *admissionregistrationv1.ValidatingWebhookConfiguration {
	obj := ValidatingWebhookConfiguration(name, vwcho...)
	mfTesting.SetManifestivalAnnotation(obj)
	mfTesting.SetLastApplied(obj)
	return obj
}

// ValidatingWebhookConfiguration creates a ValidatingWebhookConfiguration.
func ValidatingWebhookConfiguration(name string, vwcho ...ValidatingWebhookConfigurationOption) *admissionregistrationv1.ValidatingWebhookConfiguration {
	vwhc := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range vwcho {
		opt(vwhc)
	}
	return vwhc
}

// DeleteValidatingWebhookConfigurationAction creates a DeleteActionImpl that deletes
// validatingwebhookconfigurationss in Namespace test.
func DeleteValidatingWebhookConfigurationAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: "test",
			Verb:      "delete",
			Resource: schema.GroupVersionResource{
				Group:    "admissionregistration",
				Version:  "v1",
				Resource: "validatingwebhookconfigurationss",
			},
		},
		Name: name,
	}
}
