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
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeIAMPolicy(t *testing.T) {
	t.Parallel()

	// Given several spaces, each Space will have a corresponding member.
	spaceA := new(v1alpha1.Space)
	spaceA.Name = "a"
	spaceB := new(v1alpha1.Space)
	spaceB.Name = "b"
	spaceTerminating := new(v1alpha1.Space)
	spaceTerminating.Name = "terminating"
	spaceTerminating.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	spaces := []*v1alpha1.Space{spaceA, spaceB, spaceTerminating}

	policy, err := MakeIAMPolicy("some-project", "some-gsa@something.com", spaces)
	testutil.AssertErrorsEqual(t, nil, err)

	// ObjectMeta
	testutil.AssertUnstructuredEqual(t, "kind", "IAMPolicy", policy)
	testutil.AssertUnstructuredEqual(t, "apiVersion", "iam.cnrm.cloud.google.com/v1beta1", policy)
	testutil.AssertUnstructuredEqual(t, "metadata.name", "some-gsa", policy)

	// ResourceRef
	testutil.AssertUnstructuredEqual(t, "spec.resourceRef.external", "projects/some-project/serviceAccounts/some-gsa@some-project.iam.gserviceaccount.com", policy)
	testutil.AssertUnstructuredEqual(t, "spec.resourceRef.apiVersion", "iam.cnrm.cloud.google.com/v1beta1", policy)
	testutil.AssertUnstructuredEqual(t, "spec.resourceRef.kind", "IAMServiceAccount", policy)

	// Bindings
	testutil.AssertUnstructuredEqual(t, "spec.bindings", []interface{}{map[string]interface{}{
		"role": "roles/iam.workloadIdentityUser",
		"members": []interface{}{
			"serviceAccount:some-project.svc.id.goog[a/kf-builder]",
			"serviceAccount:some-project.svc.id.goog[b/kf-builder]",
			"serviceAccount:some-project.svc.id.goog[cnrm-system/cnrm-controller-manager]",
			"serviceAccount:some-project.svc.id.goog[kf/controller]",
		},
	},
	}, policy)
}

func TestMakeIAMPolicy_invalidGSA(t *testing.T) {
	t.Parallel()

	_, err := MakeIAMPolicy("some-project", "some-gsa", nil)
	testutil.AssertErrorsEqual(t, errors.New("invalid GSA"), err)
}
