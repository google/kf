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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// RoleName contains the names of ClusterRoles that Kf uses.
type RoleName string

const (
	clusterRoleKind = "ClusterRole"

	// SpaceManager holds the name of the ClusterRole for Kf Space managers.
	SpaceManager RoleName = "space-manager"
	// SpaceDeveloper holds the name of the ClusterRole for Kf Space developers.
	SpaceDeveloper RoleName = "space-developer"
	// SpaceAuditor holds the name of the ClusterRole for Kf Space auditors.
	SpaceAuditor RoleName = "space-auditor"
)

// AllRoleNames returns the list of Roles that Kf supports.
func AllRoleNames() []RoleName {
	return []RoleName{
		SpaceManager,
		SpaceDeveloper,
		SpaceAuditor,
	}
}

// RoleBindingName generates the binding name for a given Role.
func RoleBindingName(space *v1alpha1.Space, role RoleName) string {
	return v1alpha1.GenerateName(space.Name, string(role))
}

// MakeRoleBindingForClusterRole creates a blank RoleBinding for a given Role.
// These don't have subjects to ensure Kf isn't in the critical path of
// actuating or validating capability to update RoleBindings which opens the way
// to privilige escalation.
//
// Kf is responsible for ensuring this type exists, but allows users to edit it
// to perform the actual binding.
func MakeRoleBindingForClusterRole(space *v1alpha1.Space, clusterRoleName RoleName) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RoleBindingName(space, clusterRoleName),
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
			Name:     string(clusterRoleName),
		},
	}
}

// GetRoleBindingName finds the [space] RoleBinding name given an input role name.
func GetRoleBindingName(role, space string) string {
	validRoles := map[string]string{
		"SpaceManager":   v1alpha1.GenerateName(space, string(SpaceManager)),
		"SpaceDeveloper": v1alpha1.GenerateName(space, string(SpaceDeveloper)),
		"SpaceAuditor":   v1alpha1.GenerateName(space, string(SpaceAuditor)),
	}

	if roleBindingName, found := validRoles[role]; found {
		return roleBindingName
	}
	return ""
}
