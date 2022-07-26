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
	"testing"

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestFilterSubjectsByClusterRole(t *testing.T) {
	user := rbacv1.Subject{
		Kind: rbacv1.UserKind,
		Name: "manager@example.com",
	}
	group := rbacv1.Subject{
		Kind: rbacv1.GroupKind,
		Name: "developer-group@example.com",
	}
	ksa := rbacv1.Subject{
		Kind:      rbacv1.ServiceAccountKind,
		Name:      "auditor.test",
		Namespace: "test",
	}

	makeRoleBinding := func(roleName RoleName, subjects ...rbacv1.Subject) *rbacv1.RoleBinding {
		return &rbacv1.RoleBinding{
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     clusterRoleKind,
				Name:     string(roleName),
			},
			Subjects: subjects,
		}
	}

	cases := map[string]struct {
		includeRoles []RoleName
		roleBindings []*rbacv1.RoleBinding
		wantSubjects []rbacv1.Subject
	}{
		"no included roles": {
			includeRoles: nil,
			roleBindings: []*rbacv1.RoleBinding{
				makeRoleBinding(SpaceManager, user),
				makeRoleBinding(SpaceDeveloper, group),
				makeRoleBinding(SpaceAuditor, ksa),
			},
			wantSubjects: []rbacv1.Subject{},
		},
		"one matching binding": {
			includeRoles: []RoleName{SpaceManager},
			roleBindings: []*rbacv1.RoleBinding{
				makeRoleBinding(SpaceManager, user),
				makeRoleBinding(SpaceDeveloper, group),
				makeRoleBinding(SpaceAuditor, ksa),
			},
			wantSubjects: []rbacv1.Subject{user},
		},
		"only ClusterRoles match": {
			includeRoles: []RoleName{SpaceManager},
			roleBindings: (func() (out []*rbacv1.RoleBinding) {
				rb := makeRoleBinding(SpaceManager, user)
				rb.RoleRef.Kind = "Role"
				out = append(out, rb)
				return
			})(),
			wantSubjects: []rbacv1.Subject{},
		},
		"only RBAC types match": {
			includeRoles: []RoleName{SpaceManager},
			roleBindings: (func() (out []*rbacv1.RoleBinding) {
				rb := makeRoleBinding(SpaceManager, user)
				rb.RoleRef.APIGroup = "kf.dev"
				out = append(out, rb)
				return
			})(),
			wantSubjects: []rbacv1.Subject{},
		},
		"subject output is stable": {
			includeRoles: []RoleName{SpaceManager, SpaceAuditor, SpaceDeveloper},
			roleBindings: []*rbacv1.RoleBinding{
				makeRoleBinding(SpaceManager, user),
				makeRoleBinding(SpaceDeveloper, user, group),
				makeRoleBinding(SpaceAuditor, ksa, user, group),
			},
			wantSubjects: []rbacv1.Subject{group, ksa, user},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotSubjects := FilterSubjectsByClusterRole(tc.includeRoles, tc.roleBindings)

			testutil.AssertEqual(
				t,
				"subjects",
				tc.wantSubjects,
				gotSubjects,
			)
		})
	}
}

func TestMakeClusterRoleBinding(t *testing.T) {
	testSpace := &kfv1alpha1.Space{}
	testSpace.Name = "test"

	cases := map[string]struct {
		space    *kfv1alpha1.Space
		roleName RoleName
		subjects []rbacv1.Subject
	}{
		"no subjects": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceManager,
			subjects: []rbacv1.Subject{},
		},
		"nominal": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceManager,
			subjects: []rbacv1.Subject{
				{
					Kind: rbacv1.GroupKind,
					Name: "group@example.com",
				},
				{
					Kind: rbacv1.UserKind,
					Name: "someone@example.com",
				},
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "sa.test",
					Namespace: "test",
				},
			},
		},
		"duplicates": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceManager,
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.GroupKind, Name: "group@example.com"},
				{Kind: rbacv1.UserKind, Name: "someone@example.com"},
				{Kind: rbacv1.ServiceAccountKind, Name: "sa.test", Namespace: "test"},
				{Kind: rbacv1.ServiceAccountKind, Name: "sa.test", Namespace: "other"},
				{Kind: rbacv1.UserKind, Name: "someone@example.com"},
				{Kind: rbacv1.GroupKind, Name: "group@example.com"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			binding := MakeClusterRoleBinding(tc.space, tc.roleName, tc.subjects)

			testutil.AssertGoldenJSONContext(t, "clusterrolebinding", binding, map[string]interface{}{
				"space":    tc.space,
				"roleName": tc.roleName,
			})
		})
	}
}
