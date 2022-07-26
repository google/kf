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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgotesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NamespaceOption enables further configuration of a Namespace.
type NamespaceOption func(*corev1.Namespace)

// ManifestivalNamespace creates a Namespace owned by manifestival.
func ManifestivalNamespace(name string, do ...NamespaceOption) *corev1.Namespace {
	ns := Namespace(name, do...)
	mfTesting.SetManifestivalAnnotation(ns)
	mfTesting.SetLastApplied(ns)
	return ns
}

// Namespace creates a Namespace.
func Namespace(name string, do ...NamespaceOption) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range do {
		opt(ns)
	}
	return ns
}

// WithNamespaceOwnerRefs creates a NamespaceOption that updates OwnerReferences
// in Namespace.
func WithNamespaceOwnerRefs(refs ...*metav1.OwnerReference) NamespaceOption {
	return func(dep *corev1.Namespace) {
		references := dep.GetOwnerReferences()
		for _, ref := range refs {
			references = append(references, *ref)
		}
		dep.SetOwnerReferences(references)
	}
}

// DeleteNamespaceAction creates a DeleteActionImpl that deletes namespaces.
func DeleteNamespaceAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Verb: "delete",
			Resource: schema.GroupVersionResource{
				Version:  "v1",
				Resource: "namespaces",
			},
		},
		Name: name,
	}
}

// WithNamespaceAnnotation creates a NamespaceOption that sets annotation in Namespace.
func WithNamespaceAnnotation(annotations map[string]string) NamespaceOption {
	return func(ns *v1.Namespace) {
		ns.SetAnnotations(annotations)
	}
}
