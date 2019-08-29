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

package apps

import (
	"errors"
	"testing"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestConditionServiceBindingsReady(t *testing.T) {
	cases := map[string]struct {
		app    v1alpha1.App
		apiErr error

		expectDone bool
		expectErr  error
	}{
		"api error": {
			apiErr:     errors.New("api-error"),
			expectErr:  errors.New("api-error"),
			expectDone: true,
		},
		"mismatch generation": {
			app: v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: duckv1beta1.Status{
						ObservedGeneration: 3,
					},
				},
			},
			expectDone: false,
			expectErr:  nil,
		},
		"missing status": {
			app: v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: duckv1beta1.Status{
						Conditions: duckv1beta1.Conditions{},
					},
				},
			},
			expectDone: false,
		},
		"unknown status causes retry": {
			app: v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: duckv1beta1.Status{
						Conditions: duckv1beta1.Conditions{
							{Type: v1alpha1.AppConditionServiceBindingsReady, Status: "Unknown"},
						},
					},
				},
			},
			expectDone: false,
		},
		"good status completes without error": {
			app: v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: duckv1beta1.Status{
						Conditions: duckv1beta1.Conditions{
							{Type: v1alpha1.AppConditionServiceBindingsReady, Status: "True"},
						},
					},
				},
			},
			expectDone: true,
		},
		"bad status completes with error": {
			app: v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: duckv1beta1.Status{
						Conditions: duckv1beta1.Conditions{
							{Type: v1alpha1.AppConditionServiceBindingsReady, Status: "False", Message: "ShortMessage", Reason: "longer description"},
						},
					},
				},
			},
			expectDone: true,
			expectErr:  errors.New("checking ServiceBindingsReady failed, status: False message: ShortMessage reason: longer description"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualDone, actualErr := ConditionServiceBindingsReady(&tc.app, tc.apiErr)

			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "done", tc.expectDone, actualDone)
		})
	}
}
