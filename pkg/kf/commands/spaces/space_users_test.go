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

package spaces

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewSpaceUsersCommand(t *testing.T) {
	t.Parallel()
	const (
		roleNameFirst        = "role-first"
		roleBindingNameFirst = "my-space-role-first"
		roleNameLater        = "role-later"
		roleBindingNameLater = "my-space-role-later"
		spaceName            = "my-space"
		userNameFirst        = "user-first"
		userNameLater        = "user-later"
	)
	usernames := []string{userNameLater, userNameFirst}
	roleBindingFirst := createRoleBinding(roleNameFirst, roleBindingNameFirst, spaceName, usernames)
	roleBindingLater := createRoleBinding(roleNameLater, roleBindingNameLater, spaceName, usernames)
	k8sclient := k8sfake.NewSimpleClientset(&roleBindingFirst, &roleBindingLater)

	cases := map[string]struct {
		space   string
		args    []string
		wantErr error
		wantOut []string
	}{
		"wrong number of args": {
			args:    []string{"user1"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"no target Space": {
			args:    []string{},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"list users and roles in deterministic order": {
			space:   spaceName,
			args:    []string{},
			wantOut: []string{"Name        Kind  Roles\nuser-first  User  [role-first role-later]\nuser-later  User  [role-first role-later]\n"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			c := NewSpaceUsersCommand(&config.KfParams{Space: tc.space}, k8sclient)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertContainsAll(t, buffer.String(), tc.wantOut)
		})
	}
}

func createRoleBinding(roleName, roleBindingName, space string, subjects []string) rbacv1.RoleBinding {
	roleBinding := rbacv1.RoleBinding{}
	roleBinding.Name = roleBindingName
	roleBinding.Namespace = space
	roleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	roleBinding.Kind = "RoleBinding"
	roleBinding.RoleRef = rbacv1.RoleRef{
		Name: roleName,
	}
	roleBinding.Subjects = []rbacv1.Subject{}

	for _, subject := range subjects {
		roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
			Kind: rbacv1.UserKind,
			Name: subject,
		})
	}
	return roleBinding
}
