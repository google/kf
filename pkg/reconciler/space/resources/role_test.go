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
	"fmt"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	v1 "k8s.io/api/rbac/v1"
)

func ExampleAuditorRoleName() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	fmt.Println(AuditorRoleName(space))

	// Output: space-auditor
}

func ExampleDeveloperRoleName() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	fmt.Println(DeveloperRoleName(space))

	// Output: space-developer
}

func TestMakeAuditorRole(t *testing.T) {
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	ar, err := MakeAuditorRole(space)
	testutil.AssertNil(t, "MakeAuditorRole error", err)

	for _, rule := range ar.Rules {
		t.Run(fmt.Sprintf("%v/%v", rule.APIGroups, rule.Resources), func(t *testing.T) {
			testutil.AssertEqual(t, "roles are read-only", readOnlyVerbs(), rule.Verbs)
		})
	}

	// TODO(josephlewis42) fill in this table when the apps CRD gets added and all
	// of the necessary roles get finalized
	assertAllowed(t, ar, "get", "serving.knative.dev", "services")
	assertNotAllowed(t, ar, "get", "", "secrets")
}

func TestMakeDeveloperRole(t *testing.T) {
	cases := map[string]struct {
		Space  v1alpha1.Space
		Assert func(t *testing.T, role *v1.Role)
	}{
		"default space": {
			Space: v1alpha1.Space{},
			Assert: func(t *testing.T, role *v1.Role) {
				assertNotAllowed(t, role, "get", "", "pods/log")
			},
		},
		"space allows logs": {
			Space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Security: v1alpha1.SpaceSpecSecurity{
						EnableDeveloperLogsAccess: true,
					},
				},
			},
			Assert: func(t *testing.T, role *v1.Role) {
				assertAllowed(t, role, "get", "", "pods/log")
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			role, err := MakeDeveloperRole(&tc.Space)
			testutil.AssertNil(t, "MakeDeveloperRole error", err)

			tc.Assert(t, role)
		})
	}
}

func assertAllowed(t *testing.T, role *v1.Role, verb, group, resource string) {
	t.Helper()

	if policyRuleMatches(role, verb, group, resource) {
		return
	}

	t.Fatalf("no policy allowed %s on %q %q", verb, group, resource)
}

func assertNotAllowed(t *testing.T, role *v1.Role, verb, group, resource string) {
	t.Helper()

	if policyRuleMatches(role, verb, group, resource) {
		t.Fatalf("a policy allowed %s on %q %q", verb, group, resource)
		return
	}
}

func policyRuleMatches(role *v1.Role, verb, group, resource string) bool {
	for _, rule := range role.Rules {
		if listMatches(verb, rule.Verbs) &&
			listMatches(group, rule.APIGroups) &&
			listMatches(resource, rule.Resources) {
			return true
		}
	}

	return false
}

func listMatches(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == "*" || v == needle {
			return true
		}
	}

	return false
}
