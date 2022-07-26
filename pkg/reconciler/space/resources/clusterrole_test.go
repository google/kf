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

func TestMakeSpaceManagerClusterRole(t *testing.T) {
	testSpace := &kfv1alpha1.Space{}
	testSpace.Name = "test"

	cases := map[string]struct {
		space *kfv1alpha1.Space
	}{
		"nominal": {
			space: testSpace.DeepCopy(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			obj := MakeSpaceManagerClusterRole(tc.space)

			testutil.AssertGoldenJSONContext(t, "clusterrole", obj, map[string]interface{}{
				"space": tc.space,
			})
		})
	}
}
