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

package builds

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktesting "k8s.io/client-go/testing"
	duck "knative.dev/pkg/apis/duck/v1beta1"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func TestBuildStatus(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		build          v1alpha1.Build
		expectFinished bool
		expectErr      error
	}{
		"incomplete": {
			build:          v1alpha1.Build{},
			expectFinished: false,
			expectErr:      nil,
		},
		"failed": {
			build: v1alpha1.Build{
				Status: v1alpha1.BuildStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.BuildConditionSucceeded, Status: "False", Reason: "fail-reason", Message: "fail-message"},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      errors.New("build failed for reason: fail-reason with message: fail-message"),
		},
		"succeeded": {
			build: v1alpha1.Build{
				Status: v1alpha1.BuildStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.BuildConditionSucceeded, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      nil,
		},
		"still building": {
			build: v1alpha1.Build{
				Status: v1alpha1.BuildStatus{
					Status: duck.Status{
						Conditions: duck.Conditions{
							{Type: v1alpha1.BuildConditionSucceeded, Status: corev1.ConditionUnknown, Reason: "Building"},
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
			finished, err := BuildStatus(tc.build)

			testutil.AssertEqual(t, "finished", tc.expectFinished, finished)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)
		})
	}
}

func TestWaitForTaskRunPod(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		setup          func(ctx context.Context) context.Context
		ns             string
		buildName      string
		expectedErr    error
		finishDuration time.Duration
	}{
		"returns context error": {
			setup: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx
			},
			expectedErr: errors.New("context canceled"),
		},
		"dynamic client returns an error": {
			ns:        "some-ns",
			buildName: "some-build",
			setup: func(ctx context.Context) context.Context {
				fakedynamicclient.Get(ctx).
					PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some-error")
					})
				return ctx
			},
			expectedErr: errors.New("failed to get TaskRun: some-error"),
		},
		"exits once the status has a pod": {
			ns:             "some-ns",
			buildName:      "some-build",
			finishDuration: 500 * time.Millisecond,
			setup: func(ctx context.Context) context.Context {
				client := fakedynamicclient.Get(ctx).Resource(schema.GroupVersionResource{
					Group:    "tekton.dev",
					Version:  "v1beta1",
					Resource: "taskruns",
				}).Namespace("some-ns")

				// Setup with an empty podName, and then eventually set the
				// podName.
				time.AfterFunc(500*time.Millisecond, func() {
					_, err := client.Create(ctx, dynamicutils.NewUnstructured(map[string]interface{}{
						"metadata.name":      "some-build",
						"metadata.namespace": "some-ns",
						"status.podName":     "",
					}), metav1.CreateOptions{})
					testutil.AssertErrorsEqual(t, nil, err)
				})

				time.AfterFunc(1000*time.Millisecond, func() {
					_, err := client.Update(ctx, dynamicutils.NewUnstructured(map[string]interface{}{
						"metadata.name":      "some-build",
						"metadata.namespace": "some-ns",
						"status.podName":     "some-pod",
					}), metav1.UpdateOptions{})
					testutil.AssertErrorsEqual(t, nil, err)
				})

				return ctx
			},
			expectedErr: nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := fake.WithInjection(context.Background(), t)
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if tc.setup != nil {
				ctx = tc.setup(ctx)
			}

			start := time.Now()
			err := waitForTaskRunPod(ctx, tc.ns, tc.buildName, &bytes.Buffer{})
			testutil.AssertErrorsEqual(t, tc.expectedErr, err)
			testutil.AssertTrue(t, "duration", time.Since(start) >= tc.finishDuration)
		})
	}
}
