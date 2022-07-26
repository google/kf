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

package algorithms

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestSubjects_Less(t *testing.T) {
	cases := map[string]struct {
		lesser  rbacv1.Subject
		greater rbacv1.Subject
	}{
		"kind first": {
			lesser: rbacv1.Subject{
				Kind:      "AAA",
				Name:      "ZZZ",
				Namespace: "ZZZ",
			},
			greater: rbacv1.Subject{
				Kind:      "ZZZ",
				Name:      "AAA",
				Namespace: "AAA",
			},
		},
		"name second": {
			lesser: rbacv1.Subject{
				Kind:      "AAA",
				Name:      "AAA",
				Namespace: "ZZZ",
			},
			greater: rbacv1.Subject{
				Kind:      "AAA",
				Name:      "ZZZ",
				Namespace: "AAA",
			},
		},
		"namespace third": {
			lesser: rbacv1.Subject{
				Kind:      "AAA",
				Name:      "AAA",
				Namespace: "AAA",
			},
			greater: rbacv1.Subject{
				Kind:      "AAA",
				Name:      "AAA",
				Namespace: "ZZZ",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			lg := Subjects{tc.lesser, tc.greater}

			if !lg.Less(0, 1) {
				t.Error("expected lesser < greater == true")
			}

			if lg.Less(1, 0) {
				t.Error("expected greater < lesser == false")
			}

			if lg.Less(1, 1) {
				t.Error("expected greater < greater == false")
			}

			if lg.Less(0, 0) {
				t.Error("expected lesser < lesser == false")
			}
		})
	}
}

func TestSubjects_Contains(t *testing.T) {
	cases := map[string]struct {
		Subjects  []rbacv1.Subject
		Name      string
		Kind      string
		Contained bool
		Index     int
	}{
		"subject not contained": {
			Subjects: []rbacv1.Subject{
				{
					Name: "user1",
					Kind: "User",
				},
			},
			Name:      "user2",
			Kind:      "User",
			Contained: false,
			Index:     -1,
		},
		"subject is contained": {
			Subjects: []rbacv1.Subject{
				{
					Name: "user1",
					Kind: "User",
				},
			},
			Name:      "user1",
			Kind:      "User",
			Contained: true,
			Index:     0,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			subjects := Subjects(tc.Subjects)

			contained, index := subjects.Contains(tc.Name, tc.Kind)

			testutil.AssertEqual(t, "contained", contained, tc.Contained)
			testutil.AssertEqual(t, "index", index, tc.Index)
		})
	}
}
