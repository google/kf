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

package resources

import (
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/algorithms"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/kmeta"
)

const (
	// ClusterReaderRole holds the name of the ClusterRole that grants read access
	// at the cluster scope for Kf developers, managers, and auditors.
	ClusterReaderRole RoleName = "kf-cluster-reader"
)

// ClusterRoleBindingName generates the binding name for a given Role.
func ClusterRoleBindingName(space *v1alpha1.Space, role RoleName) string {
	return v1alpha1.GenerateName(space.Name, string(role))
}

// FilterSubjectsByClusterRole returns a list of subjects from the given RoleBinding
// that have any of the given ClusterRoles.
//
// The returned list will be in a determinstic order.
func FilterSubjectsByClusterRole(includeRoles []RoleName, roleBindings []*rbacv1.RoleBinding) []rbacv1.Subject {
	allowedRoles := sets.NewString()
	for _, role := range includeRoles {
		allowedRoles.Insert(string(role))
	}

	var subjects []rbacv1.Subject
	for _, binding := range roleBindings {
		// Check that the binding refers to a supported ClusterRole.
		if rr := binding.RoleRef; rr.APIGroup != rbacv1.GroupName ||
			rr.Kind != clusterRoleKind ||
			!allowedRoles.Has(rr.Name) {
			continue
		}

		for _, sub := range binding.Subjects {
			subjects = append(subjects, *sub.DeepCopy())
		}
	}

	return algorithms.Dedupe(
		algorithms.Subjects(subjects),
	).(algorithms.Subjects)
}

// MakeClusterRoleBinding creates a populated RoleBinding for a given Role.
func MakeClusterRoleBinding(space *v1alpha1.Space, role RoleName, subjects []rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	deduplicatedSubjects := algorithms.Dedupe(
		algorithms.Subjects(subjects),
	).(algorithms.Subjects)

	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterRoleBindingName(space, role),
			Namespace: NamespaceName(space),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			// Copy labels from the parent.
			Labels: v1alpha1.UnionMaps(
				space.GetLabels(),
				map[string]string{
					managedByLabel: "kf",
				}),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     clusterRoleKind,
			Name:     string(role),
		},
		Subjects: deduplicatedSubjects,
	}
}
