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
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/testutil"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewSetSpaceRoleCommand(t *testing.T) {
	t.Parallel()

	const (
		roleBindingName = "my-space-space-manager"
		spaceName       = "my-space"
	)
	roleBinding := &rbacv1.RoleBinding{}
	roleBinding.Name = roleBindingName
	roleBinding.Namespace = spaceName
	roleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	roleBinding.Kind = "RoleBinding"
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind: rbacv1.UserKind,
			Name: "existed_user",
		},
	}

	k8sclient := k8sfake.NewSimpleClientset(roleBinding)

	cases := map[string]struct {
		space           string
		wantErr         error
		args            []string
		expectedStrings []string
	}{
		"0 args": {
			args:    []string{},
			wantErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"1 arg": {
			args:    []string{"user1"},
			wantErr: errors.New("accepts 2 arg(s), received 1"),
		},
		"wrong number of args": {
			args:    []string{"user1", "role1", "role2"},
			wantErr: errors.New("accepts 2 arg(s), received 3"),
		},
		"no target Space": {
			args:    []string{"user1", "role1"},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"invalid role name": {
			space:   spaceName,
			args:    []string{"user1", "role1"},
			wantErr: errors.New("Role \"role1\" does not exist"),
		},
		"user is already assigned role": {
			space:           spaceName,
			args:            []string{"existed_user", "SpaceManager"},
			expectedStrings: []string{"\"existed_user\" (\"User\") is already assigned \"SpaceManager\" Role"},
		},
		"user is assigned role": {
			space:           spaceName,
			args:            []string{"user-a", "SpaceManager"},
			expectedStrings: []string{"\"user-a\" (User) is assigned \"SpaceManager\" Role"},
		},
		"Group is assigned role": {
			space:           spaceName,
			args:            []string{"group-A", "SpaceManager", "-t", "Group"},
			expectedStrings: []string{"\"group-A\" (Group) is assigned \"SpaceManager\" Role"},
		},
		"ServiceAccount is assigned role": {
			space:           spaceName,
			args:            []string{"SA1", "SpaceManager", "-t", "ServiceAccount"},
			expectedStrings: []string{"\"SA1\" (ServiceAccount) is assigned \"SpaceManager\" Role"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gomock.NewController(t)
			buffer := new(bytes.Buffer)

			ctx := configlogging.SetupLogger(context.Background(), buffer)

			c := NewSetSpaceRoleCommand(&config.KfParams{Space: tc.space}, k8sclient)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			c.SetContext(ctx)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertContainsAll(t, buffer.String(), tc.expectedStrings)

		})
	}
}
