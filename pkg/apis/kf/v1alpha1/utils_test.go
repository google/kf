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
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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
			expectMessage: "status not yet reconciled",
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
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")

	err := scm.MarkChildNotOwned("sample-123")

	result := manager.GetCondition("SampleReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: There is an existing Sample "sample-123" that we do not own.
	// Reason: NotOwned
	// Error: There is an existing Sample "sample-123" that we do not own.
}

func ExampleSingleConditionManager_MarkTemplateError() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")

	err := scm.MarkTemplateError(errors.New("tmpl-err"))

	result := manager.GetCondition("SampleReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: Couldn't populate the Sample template: tmpl-err
	// Reason: TemplateError
	// Error: Couldn't populate the Sample template: tmpl-err
}

func ExampleSingleConditionManager_MarkReconciliationError() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")

	err := scm.MarkReconciliationError("updating", errors.New("update err"))

	result := manager.GetCondition("SampleReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: False
	// Message: Error occurred while updating Sample: update err
	// Reason: ReconciliationError
	// Error: Error occurred while updating Sample: update err
}

func ExampleSingleConditionManager_MarkReconciliationError_conflict() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")

	conflict := apierrs.NewConflict(
		schema.GroupResource{
			Group:    "serving.knative.dev",
			Resource: "services",
		},
		"MyService",
		errors.New("the object has been modified; please apply your changes to the latest version and try again"),
	)
	err := scm.MarkReconciliationError("updating", conflict)

	result := manager.GetCondition("SampleReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)
	fmt.Println("Error:", err)

	// Output: Status: Unknown
	// Message: Error occurred while updating Sample: Operation cannot be fulfilled on services.serving.knative.dev "MyService": the object has been modified; please apply your changes to the latest version and try again
	// Reason: CacheOutdated
	// Error: Error occurred while updating Sample: Operation cannot be fulfilled on services.serving.knative.dev "MyService": the object has been modified; please apply your changes to the latest version and try again
}

func ExampleSingleConditionManager_MarkUnknown() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")

	scm.MarkUnknown("FetchError", "couldn't fetch value: %v", errors.New("some err"))

	result := manager.GetCondition("SampleReady")
	fmt.Println("Status:", result.Status)
	fmt.Println("Message:", result.Message)
	fmt.Println("Reason:", result.Reason)

	// Output: Status: Unknown
	// Message: couldn't fetch value: some err
	// Reason: FetchError
}

func TestGenerateName(t *testing.T) {
	cases := map[string]struct {
		input          []string
		expectedOutput string
	}{
		"short": {
			input:          []string{"my-app"},
			expectedOutput: "my-app",
		},
		"multiple entries": {
			input:          []string{"a", "b", "c"},
			expectedOutput: "a-b-c",
		},
		"invalid prefix/suffix chars": {
			input:          []string{"---abc---"},
			expectedOutput: "abc75b6fb067496dbcc888ccce24201eddd",
		},
		"too long": {
			input: []string{
				"integration-push-spring-music",
				"1582060202119043190",
				"ro3jy6g1eheb",
				"38n6dgwh9kx7h",
				"c0pljctl1vq7",
			},
			expectedOutput: "integration-push-spring-music-1b532b669f9f4360daad57088be0aee0d",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			output := GenerateName(tc.input...)
			testutil.AssertEqual(t, "output", tc.expectedOutput, output)

			// Kubernetes names and labels must be <= 63 characters
			testutil.AssertTrue(t, "length <= 63", len(output) <= 63)
		})
	}
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
		testutil.AssertEqual(t, fmt.Sprintf("len: %d", len(r)), true, len(r) < 64)
		testutil.AssertEqual(t, "collison", false, history[r])
	}

	for i := 0; i < 10000; i++ {
		r := GenerateName(randStr(), randStr())
		validDNS(r)
		history[r] = true
	}
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

func ExampleManagedByKfRequirement() {
	requirement := ManagedByKfRequirement()
	selector := labels.NewSelector().Add(requirement).String()
	fmt.Println(selector)

	// Output: app.kubernetes.io/managed-by=kf
}

func TestSummarizeChildConditions(t *testing.T) {
	cases := map[string]struct {
		conditions  []apis.Condition
		wantOverall apis.Condition
	}{
		"empty conditions": {
			conditions: []apis.Condition{},
			wantOverall: apis.Condition{
				Type:   apis.ConditionReady,
				Status: corev1.ConditionTrue,
			},
		},
		"nil conditions": {
			conditions: nil,
			wantOverall: apis.Condition{
				Type:   apis.ConditionReady,
				Status: corev1.ConditionTrue,
			},
		},
		"ready false": {
			conditions: []apis.Condition{
				{Type: "subresource1", Status: corev1.ConditionFalse, Reason: "reason", Message: "message"},
			},
			wantOverall: apis.Condition{
				Type:    apis.ConditionReady,
				Status:  corev1.ConditionFalse,
				Reason:  "reason",
				Message: "message",
			},
		},
		"ready unknown": {
			conditions: []apis.Condition{
				{Type: "subresource1", Status: corev1.ConditionUnknown, Reason: "reason", Message: "message"},
			},
			wantOverall: apis.Condition{
				Type:    apis.ConditionReady,
				Status:  corev1.ConditionUnknown,
				Reason:  "reason",
				Message: "message",
			},
		},
		"ready true": {
			conditions: []apis.Condition{
				{Type: "subresource1", Status: corev1.ConditionTrue, Reason: "reason", Message: "message"},
			},
			wantOverall: apis.Condition{
				Type:   apis.ConditionReady,
				Status: corev1.ConditionTrue,
				// TODO: support reasons/messages for True conditions when Knative does
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotOverall, _ := SummarizeChildConditions(tc.conditions)

			testutil.AssertNotNil(t, "overall", gotOverall)
			// remove time because it's nondeterministic
			gotOverall.LastTransitionTime = apis.VolatileTime{}
			testutil.AssertEqual(t, "overall", tc.wantOverall, *gotOverall)
		})
	}
}

func TestCopyMap(t *testing.T) {
	cases := map[string]struct {
		input          map[string]string
		expectedOutput map[string]string
	}{
		"nil": {
			input:          nil,
			expectedOutput: nil,
		},
		"empty": {
			input:          map[string]string{},
			expectedOutput: map[string]string{},
		},
		"valid": {
			input: map[string]string{
				"disktype": "ssd",
				"cpu":      "amd64",
			},
			expectedOutput: map[string]string{
				"disktype": "ssd",
				"cpu":      "amd64",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			output := CopyMap(tc.input)
			testutil.AssertEqual(t, "output", tc.expectedOutput, output)
		})
	}
}

func ExampleSingleConditionManager_String() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")
	scm.MarkFalse("BadState", "Some message here")

	fmt.Println(scm.String())

	// Output: condition: SampleReady status: "False" reason: "BadState" message: "Some message here"
}

func ExampleSingleConditionManager_TimeSinceTransition() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})

	// Force a state transition to initialize the time
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")
	scm.MarkFalse("BadState", "Some message here")

	a := scm.TimeSinceTransition()
	time.Sleep(100 * time.Millisecond)
	b := scm.TimeSinceTransition()
	fmt.Println("Increasing Time:", b > a)

	// Output: Increasing Time: true
}

func ExampleSingleConditionManager_ErrorIfTimeout() {
	manager := apis.NewLivingConditionSet("SampleReady").Manage(&duckv1beta1.Status{})

	// Force a state transition to initialize the time
	scm := NewSingleConditionManager(manager, "SampleReady", "Sample")
	scm.MarkUnknown("UnknownState", "Some message here")

	fmt.Println("Timeout not reached err:", scm.ErrorIfTimeout(1*time.Hour))
	fmt.Println("Timeout reached err:", scm.ErrorIfTimeout(-1*time.Second))

	// Output: Timeout not reached err: <nil>
	// Timeout reached err: timed out, no progress was made in -1 seconds, previous status: "condition: SampleReady status: \"Unknown\" reason: \"UnknownState\" message: \"Some message here\""
}
