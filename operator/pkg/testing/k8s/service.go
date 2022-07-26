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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// ServiceOption enables further configuration of a Service.
type ServiceOption func(*corev1.Service)

// ManifestivalService creates a Service owned by manifestival.
func ManifestivalService(name string, so ...ServiceOption) *corev1.Service {
	obj := Service(name, so...)
	mfTesting.SetManifestivalAnnotation(obj)
	mfTesting.SetLastApplied(obj)
	return obj
}

// Service creates a Service.
func Service(name string, so ...ServiceOption) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range so {
		opt(svc)
	}
	return svc
}

// WithNamespace configures the namespace of the service.
func WithNamespace(namespace string) ServiceOption {
	return func(svc *corev1.Service) {
		svc.Namespace = namespace
	}
}

// WithLabelSelector configures the selector of the service.
func WithLabelSelector(selector map[string]string) ServiceOption {
	return func(svc *corev1.Service) {
		svc.Spec.Selector = selector
	}
}

// DeleteServiceAction creates a DeleteActionImpl that deletes services
// in Namespace test.
func DeleteServiceAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: "test",
			Verb:      "delete",
			Resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1",
				Resource: "services",
			},
		},
		Name: name,
	}
}
