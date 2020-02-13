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

package v1alpha1

import (
	"errors"
	"fmt"
	"math/rand"
	"path"
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestPropagateCondition(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		source        *apis.Condition
		expectMessage string
		expectStatus  string
		expectReason  string
		expectReturn  bool
	}{
		"nil source": {
			source:        nil,
			expectMessage: "source status is nil",
			expectStatus:  "Unknown",
			expectReason:  "Unknown",
			expectReturn:  false,
		},
		"unknown source": {
			source:        &apis.Condition{Status: "Unknown", Message: "u-message", Reason: "UReason"},
			expectMessage: "u-message",
			expectStatus:  "Unknown",
			expectReason:  "UReason",
			expectReturn:  false,
		},
		"false source": {
			source:        &apis.Condition{Status: "False", Message: "f-message", Reason: "FReason"},
			expectMessage: "f-message",
			expectStatus:  "False",
			expectReason:  "FReason",
			expectReturn:  false,
		},
		"true source": {
			// messages and reasons are excluded from success, even if they're present
			// on the child
			source:        &apis.Condition{Status: "True", Message: "t-message", Reason: "TReason"},
			expectMessage: "",
			expectStatus:  "True",
			expectReason:  "",
			expectReturn:  true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := &duckv1beta1.Status{}
			manager := apis.NewLivingConditionSet("TestCondition").Manage(status)

			returnValue := PropagateCondition(manager, "TestCondition", tc.source)
			testutil.AssertEqual(t, "return value", tc.expectReturn, returnValue)

			resultCond := manager.GetCondition("TestCondition")
			testutil.AssertEqual(t, "message", tc.expectMessage, resultCond.Message)
			testutil.AssertEqual(t, "status", tc.expectStatus, string(resultCond.Status))
			testutil.AssertEqual(t, "reason", tc.expectReason, resultCond.Reason)
		})
	}
}

func ExampleSingleConditionManager_MarkChildNotOwned() {
	manager := apis.NewLivingConditionSet("DummyReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "DummyReady", "Dummy")

	err := scm.MarkChildNotOwned("dummy-123")

	result := manager.GetCondition("DummyReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: There is an existing Dummy "dummy-123" that we do not own.
	// Reason: NotOwned
	// Error: There is an existing Dummy "dummy-123" that we do not own.
}

func ExampleSingleConditionManager_MarkTemplateError() {
	manager := apis.NewLivingConditionSet("DummyReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "DummyReady", "Dummy")

	err := scm.MarkTemplateError(errors.New("tmpl-err"))

	result := manager.GetCondition("DummyReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: Couldn't populate the Dummy template: tmpl-err
	// Reason: TemplateError
	// Error: Couldn't populate the Dummy template: tmpl-err
}

func ExampleSingleConditionManager_MarkReconciliationError() {
	manager := apis.NewLivingConditionSet("DummyReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "DummyReady", "Dummy")

	err := scm.MarkReconciliationError("updating", errors.New("update err"))

	result := manager.GetCondition("DummyReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: Error occurred while updating Dummy: update err
	// Reason: ReconciliationError
	// Error: Error occurred while updating Dummy: update err
}

func ExampleSingleConditionManager_MarkReconciliationError_conflict() {
	manager := apis.NewLivingConditionSet("DummyReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "DummyReady", "Dummy")

	conflict := apierrs.NewConflict(
		schema.GroupResource{
			Group:    "serving.knative.dev",
			Resource: "services",
		},
		"MyService",
		errors.New("the object has been modified; please apply your changes to the latest version and try again"),
	)
	err := scm.MarkReconciliationError("updating", conflict)

	result := manager.GetCondition("DummyReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: Unknown
	// Message: Error occurred while updating Dummy: Operation cannot be fulfilled on services.serving.knative.dev "MyService": the object has been modified; please apply your changes to the latest version and try again
	// Reason: CacheOutdated
	// Error: Error occurred while updating Dummy: Operation cannot be fulfilled on services.serving.knative.dev "MyService": the object has been modified; please apply your changes to the latest version and try again
}

func TestGenerateName_Deterministic(t *testing.T) {
	t.Parallel()

	r1 := GenerateName("host-1", "example1.com")
	r2 := GenerateName("host-1", "example1.com")
	r3 := GenerateName("host-2", "example1.com")
	r4 := GenerateName("host-1", "example2.com")

	testutil.AssertEqual(t, "r1 and r2", r1, r2)
	testutil.AssertEqual(t, "r1 and r2", r1, r2)

	for _, r := range []string{r3, r4} {
		if r1 == r {
			t.Fatalf("expected %s to not equal %s", r, r1)
		}
	}
}

func TestGenerateName_ValidDNS(t *testing.T) {
	t.Parallel()

	// We'll use an instantiation of rand so we can seed it with 0 for
	// repeatable tests.
	rand := rand.New(rand.NewSource(0))
	randStr := func() string {
		buf := make([]byte, rand.Intn(128)+1)
		for i := range buf {
			buf[i] = byte(rand.Intn('z'-'a') + 'a')
		}
		return strings.ToUpper(path.Join("./", string(buf)))
	}

	history := map[string]bool{}

	validDNS := func(r string) {
		testutil.AssertRegexp(t, "valid DNS", `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`, r)
		testutil.AssertEqual(t, fmt.Sprintf("len: %d", len(r)), true, len(r) <= 64)
		testutil.AssertEqual(t, "collison", false, history[r])
	}

	for i := 0; i < 10000; i++ {
		r := GenerateName(randStr(), randStr())
		validDNS(r)
		history[r] = true
	}

	// Empty name
	validDNS(GenerateName())

	// Only non-alphanumeric characters
	validDNS(GenerateName(".", "-", "$"))
}

func TestIsStatusFinal(t *testing.T) {
	cases := map[string]struct {
		condition apis.Condition

		wantFinal bool
	}{
		"ready unknown": {
			condition: apis.Condition{Type: apis.ConditionReady, Status: corev1.ConditionUnknown},
			wantFinal: false,
		},
		"ready true": {
			condition: apis.Condition{Type: apis.ConditionReady, Status: corev1.ConditionTrue},
			wantFinal: true,
		},
		"ready false": {
			condition: apis.Condition{Type: apis.ConditionReady, Status: corev1.ConditionFalse},
			wantFinal: true,
		},
		"succeeded unknown": {
			condition: apis.Condition{Type: apis.ConditionSucceeded, Status: corev1.ConditionUnknown},
			wantFinal: false,
		},
		"succeeded true": {
			condition: apis.Condition{Type: apis.ConditionSucceeded, Status: corev1.ConditionTrue},
			wantFinal: true,
		},
		"succeeded false": {
			condition: apis.Condition{Type: apis.ConditionSucceeded, Status: corev1.ConditionFalse},
			wantFinal: true,
		},
		"unknown type unknown": {
			condition: apis.Condition{Type: "SecretStatus", Status: corev1.ConditionUnknown},
			wantFinal: false,
		},
		"unknown type true": {
			condition: apis.Condition{Type: "SecretStatus", Status: corev1.ConditionTrue},
			wantFinal: false,
		},
		"unknown type false": {
			condition: apis.Condition{Type: "SecretStatus", Status: corev1.ConditionFalse},
			wantFinal: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			duck := duckv1beta1.Status{Conditions: duckv1beta1.Conditions{tc.condition}}
			actualFinal := IsStatusFinal(duck)

			testutil.AssertEqual(t, "final", tc.wantFinal, actualFinal)

		})
	}
}

func ExampleUnionMaps() {
	x := map[string]string{"a": "1", "b": "x", "c": "x"}
	y := map[string]string{"a": "1", "b": "2", "c": "3"}
	z := map[string]string{"a": "1", "b": "2", "d": "4"}

	result := UnionMaps(x, y, z)

	for _, key := range []string{"a", "b", "c", "d"} {
		fmt.Printf("%s: %s\n", key, result[key])
	}

	// Output: a: 1
	// b: 2
	// c: 3
	// d: 4
}
