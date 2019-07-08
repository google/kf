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
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestPropagateCondition(t *testing.T) {
	cases := map[string]struct {
		source        *apis.Condition
		expectMessage string
		expectStatus  string
		expectReason  string
		expectReturn  bool
	}{
		"nil source": {
			source: nil,
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
			if tc.source == nil {
				testutil.AssertEqual(t, "condition", (*apis.Condition)(nil), resultCond)
				return
			}

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
