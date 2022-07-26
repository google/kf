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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MakeIAMPolicy returns an IAM Policy for the given Spaces.
func MakeIAMPolicy(project, gsaName string, spaces []*v1alpha1.Space) (*unstructured.Unstructured, error) {
	// Extract the username from the GSA email address.
	parts := strings.Split(gsaName, "@")
	if len(parts) != 2 {
		return nil, errors.New("invalid GSA")
	}

	username := parts[0]

	// NOTE: The type needs to be []interface{} instead of []string because
	// the value back from K8s will have this type. This implies that when the
	// equality detectors look at them, the type must match.
	members := []interface{}{
		fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager]", project),
		fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/controller]", project, v1alpha1.KfNamespace),
	}
	for _, space := range spaces {
		// Don't include Spaces that are terminating.
		if space.DeletionTimestamp != nil {
			continue
		}

		members = append(members, fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/kf-builder]", project, space.Name))
	}

	// Ensure it is deterministic.
	sort.Slice(members, func(i, j int) bool {
		a, b := members[i].(string), members[j].(string)
		return a < b
	})

	external := fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com", project, username, project)

	return dynamicutils.NewUnstructured(
		map[string]interface{}{
			"kind":                        "IAMPolicy",
			"apiVersion":                  "iam.cnrm.cloud.google.com/v1beta1",
			"metadata.name":               username,
			"spec.resourceRef.external":   external,
			"spec.resourceRef.apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
			"spec.resourceRef.kind":       "IAMServiceAccount",
			"spec.bindings": []interface{}{
				map[string]interface{}{
					"role":    "roles/iam.workloadIdentityUser",
					"members": members,
				},
			},
		},
	), nil
}
