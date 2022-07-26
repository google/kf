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
)

func TestMakeRoleBindingForClusterRole(t *testing.T) {
	testSpace := &kfv1alpha1.Space{}
	testSpace.Name = "test"

	cases := map[string]struct {
		space    *kfv1alpha1.Space
		roleName RoleName
	}{
		"manager": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceManager,
		},
		"auditor": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceAuditor,
		},
		"developer": {
			space:    testSpace.DeepCopy(),
			roleName: SpaceDeveloper,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			svc := MakeRoleBindingForClusterRole(tc.space, tc.roleName)

			testutil.AssertGoldenJSONContext(t, "rolebinding", svc, map[string]interface{}{
				"space":    tc.space,
				"roleName": tc.roleName,
			})
		})
	}
}

func TestGetRoleBindingName(t *testing.T) {
	cases := map[string]struct {
		Role                    string
		SpaceName               string
		ExpectedRoleBindingName string
	}{
		"role found": {
			Role:                    "SpaceManager",
			SpaceName:               "my-space",
			ExpectedRoleBindingName: "my-space-space-manager",
		},
		"role not found": {
			Role:                    "InvalidRole",
			SpaceName:               "my-space",
			ExpectedRoleBindingName: "",
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			roleBindingName := GetRoleBindingName(tc.Role, tc.SpaceName)
			testutil.AssertEqual(t, "test", roleBindingName, tc.ExpectedRoleBindingName)
		})
	}
}
