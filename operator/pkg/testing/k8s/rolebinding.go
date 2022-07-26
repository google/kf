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

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// RoleBindingOption enables further configuration of a RoleBinding.
type RoleBindingOption func(*rbacv1.RoleBinding)

// ManifestivalRoleBinding creates a RoleBinding owned by manifestival.
func ManifestivalRoleBinding(name string, rbo ...RoleBindingOption) *rbacv1.RoleBinding {
	obj := RoleBinding(name, rbo...)
	mfTesting.SetManifestivalAnnotation(obj)
	mfTesting.SetLastApplied(obj)
	return obj
}

// RoleBinding creates a RoleBinding.
func RoleBinding(name string, rbo ...RoleBindingOption) *rbacv1.RoleBinding {
	rolebinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range rbo {
		opt(rolebinding)
	}
	return rolebinding
}

// DeleteRoleBindingAction creates a DeleteActionImpl that deletes rolebindings
// in Namespace test.
func DeleteRoleBindingAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: "test",
			Verb:      "delete",
			Resource: schema.GroupVersionResource{
				Group:    "rbac",
				Version:  "v1",
				Resource: "rolebindings",
			},
		},
		Name: name,
	}
}
