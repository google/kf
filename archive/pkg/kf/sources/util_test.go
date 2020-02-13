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

package sources_test

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/sources"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	duck "knative.dev/pkg/apis/duck/v1beta1"
)

func TestBuildStatus(t *testing.T) {
	cases := map[string]struct {
		source         v1alpha1.Source
		expectFinished bool
		expectErr      error
	}{
		"incomplete": {
			source:         v1alpha1.Source{},
			expectFinished: false,
			expectErr:      nil,
		},
		"failed": {
			source: v1alpha1.Source{
				Status: v1alpha1.SourceStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.SourceConditionSucceeded, Status: "False", Reason: "fail-reason", Message: "fail-message"},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      errors.New("build failed for reason: fail-reason with message: fail-message"),
		},
		"succeeded": {
			source: v1alpha1.Source{
				Status: v1alpha1.SourceStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.SourceConditionSucceeded, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      nil,
		},
		"still building": {
			source: v1alpha1.Source{
				Status: v1alpha1.SourceStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.SourceConditionSucceeded, Status: corev1.ConditionUnknown, Reason: "Building"},
						},
					},
				},
			},
			expectFinished: false,
			expectErr:      nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			finished, err := sources.SourceStatus(tc.source)

			testutil.AssertEqual(t, "finished", tc.expectFinished, finished)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)
		})
	}
}
