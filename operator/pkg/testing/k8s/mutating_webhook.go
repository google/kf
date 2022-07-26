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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MutatingWebhookConfigurationOption enables further configuration of a MutatingWebhookConfiguration.
type MutatingWebhookConfigurationOption func(*admissionregistrationv1.MutatingWebhookConfiguration)

// MutatingWebhookConfiguration creates a MutatingWebhookConfiguration
// and then applies MutatingWebhookConfigurationOptions to it.
func MutatingWebhookConfiguration(name string, mwcho ...MutatingWebhookConfigurationOption) *admissionregistrationv1.MutatingWebhookConfiguration {
	mwhc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range mwcho {
		opt(mwhc)
	}
	return mwhc
}
