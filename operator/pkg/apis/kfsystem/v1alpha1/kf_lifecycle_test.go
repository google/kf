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

package v1alpha1

import (
	"testing"

	apistest "knative.dev/pkg/apis/testing"
)

func TestIsEnabled(t *testing.T) {
	testCases := []struct {
		name     string
		spec     *KfSpec
		expected bool
	}{
		{
			name: "nil enable",
			spec: &KfSpec{
				Enabled: nil,
			},
			// We want to ensure that even when enable isnt' set, it's still
			// enabled.
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual, expected := tc.spec.IsEnabled(), tc.expected; actual != expected {
				t.Fatalf("expected %v, got %v", expected, actual)
			}
		})
	}
}

func TestMarkKfInstallSucceeded(t *testing.T) {
	kfs := &KfSystemStatus{}
	kfs.InitializeConditions()
	kfs.MarkKfInstallSucceeded("1.2.3")

	apistest.CheckConditionSucceeded(kfs, KfInstallSucceeded, t)
	if got, want := kfs.KfVersion, "1.2.3"; got != want {
		t.Errorf("Version: got %v, want %v", got, want)
	}
}

func TestMarkKfInstallFailed(t *testing.T) {
	kfs := &KfSystemStatus{}
	kfs.InitializeConditions()
	kfs.MarkKfInstallFailed("test")

	apistest.CheckConditionFailed(kfs, KfInstallSucceeded, t)
	s := kfs.GetCondition(KfInstallSucceeded)
	if got, want := s.Reason, "Error"; got != want {
		t.Errorf("Condition Reason: got %v, want %v", got, want)
	}
}

func TestMarkKfInstallNotReady(t *testing.T) {
	kfs := &KfSystemStatus{}
	kfs.InitializeConditions()
	kfs.MarkKfInstallNotReady()

	apistest.CheckConditionOngoing(kfs, KfInstallSucceeded, t)
	s := kfs.GetCondition(KfInstallSucceeded)
	if got, want := s.Reason, "NotReady"; got != want {
		t.Errorf("Condition Reason: got %v, want %v", got, want)
	}
}
