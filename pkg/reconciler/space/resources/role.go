// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/pkg/kmeta"
	"github.com/knative/serving/pkg/resources"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeveloperRoleName gets the name of the developer role given the space.
func DeveloperRoleName(space *v1alpha1.Space) string {
	return "space-developer"
}

// MakeDeveloperRole creates a Role for developer access from a Space object.
func MakeDeveloperRole(space *v1alpha1.Space) (*v1.Role, error) {
	return &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeveloperRoleName(space),
			Namespace: NamespaceName(space),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			Labels: resources.UnionMaps(space.GetLabels(), map[string]string{
				"app.kubernetes.io/managed-by": "kf",
			}),
		},
		Rules: developerPolicyRules(space),
	}, nil
}

// AuditorRoleName gets the name of the auditor role given the space.
func AuditorRoleName(space *v1alpha1.Space) string {
	return "space-auditor"
}

// MakeAuditorRole creates a Role for auditor access from a Space object.
func MakeAuditorRole(space *v1alpha1.Space) (*v1.Role, error) {
	return &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditorRoleName(space),
			Namespace: NamespaceName(space),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			Labels: resources.UnionMaps(space.GetLabels(), map[string]string{
				managedByLabel: "kf",
			}),
		},
		Rules: auditPolicyRules(space),
	}, nil
}

func readOnlyVerbs() []string {
	return []string{"get", "list", "watch"}
}

func editVerbs() []string {
	return []string{"create", "update", "patch", "delete"}
}

func readEditVerbs() []string {
	return append(readOnlyVerbs(), editVerbs()...)
}

func developerPolicyRules(space *v1alpha1.Space) []v1.PolicyRule {
	modifyRules := []v1.PolicyRule{
		// Create service instances and bindings
		{
			APIGroups: []string{"servicecatalog.k8s.io"},
			Verbs:     editVerbs(),
			Resources: []string{
				"serviceinstances",
				"servicebindings",
			},
		},
		// Create and modify secrets
		{
			APIGroups: []string{""}, // "" is the builtin API group
			Verbs:     readEditVerbs(),
			Resources: []string{"secrets"},
		},
		// Create new services
		{
			APIGroups: []string{"serving.knative.dev"},
			Verbs:     editVerbs(),
			Resources: []string{"services"},
		},
	}

	out := append(auditPolicyRules(space), modifyRules...)

	if space.Spec.Security.EnableDeveloperLogsAccess {
		// Taken from: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
		out = append(out, v1.PolicyRule{
			APIGroups: []string{""}, // "" is the builtin API group
			Verbs:     readOnlyVerbs(),
			Resources: []string{"pods", "pods/log"},
		})
	}

	return out
}

func auditPolicyRules(space *v1alpha1.Space) []v1.PolicyRule {
	return []v1.PolicyRule{
		// Read access to Knative serving.
		{
			APIGroups: []string{"serving.knative.dev"},
			Verbs:     readOnlyVerbs(),
			Resources: []string{"*"},
		},
		// Read access to Knative build.
		{
			APIGroups: []string{"build.knative.dev"},
			Verbs:     readOnlyVerbs(),
			Resources: []string{"*"},
		},
		// Read access to the Service catalog.
		{
			APIGroups: []string{"servicecatalog.k8s.io"},
			Verbs:     readOnlyVerbs(),
			Resources: []string{"*"},
		},
		// Read access to builtins.
		{
			APIGroups: []string{""}, // "" is the builtin API group
			Verbs:     readOnlyVerbs(),
			Resources: []string{"pods", "resourcequotas", "services"},
		},
	}
}
