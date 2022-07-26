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

const (
	proxyApiVerb   = "proxy"
	uploadApiGroup = "upload.kf.dev"
)

func buildRoleName(space *v1alpha1.Space) string {
	return v1alpha1.GenerateName(space.Name, "source-builder")
}

// MakeSourceBuilderRole creates a Role to allow requests to the sourcepackages
// upload subresource api.
func MakeSourceBuilderRole(space *v1alpha1.Space) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildRoleName(space),
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

		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{proxyApiVerb},
				APIGroups: []string{uploadApiGroup},
				Resources: []string{"*"},
			},
		},
	}
}
