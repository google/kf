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

// ClusterRoleName generates the ClusterRole name for a given Space.
func ClusterRoleName(space *v1alpha1.Space) RoleName {
	return RoleName(v1alpha1.GenerateName(space.Name, "manager"))
}

// MakeSpaceManagerClusterRole creates a ClusterRole that gives Space managers
// the ability to read and modify the given Space (but not create or delete it).
func MakeSpaceManagerClusterRole(space *v1alpha1.Space) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(ClusterRoleName(space)),
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

		Rules: []rbacv1.PolicyRule{
			{
				// Creation/deletion isn't allowed for Space managers.
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
				APIGroups: []string{"kf.dev"},
				Resources: []string{"spaces"},
				// Only apply to the given Space.
				ResourceNames: []string{space.Name},
			},
		},
	}
}
