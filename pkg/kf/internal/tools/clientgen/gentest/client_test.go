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

package gentest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	testclient "k8s.io/client-go/kubernetes/fake"
	cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Namespace string
		Name      string
		ExpectErr error
		Setup     func(mockK8s cv1.PodsGetter)
	}{
		"object does not exist": {
			Namespace: "default",
			Name:      "some-object",
			ExpectErr: errors.New(`couldn't delete the OperatorConfig with the name "some-object": pods "some-object" not found`),
		},
		"object exists": {
			Namespace: "my-namespace",
			Name:      "some-object",
			Setup: func(mockK8s cv1.PodsGetter) {
				obj := &v1.Pod{}
				obj.Name = "some-object"
				mockK8s.Pods("my-namespace").Create(context.Background(), obj, metav1.CreateOptions{})
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset().CoreV1()
			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			client := coreClient{
				kclient: mockK8s,
			}

			actualErr := client.Delete(context.Background(), tc.Namespace, tc.Name)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			objects, err := mockK8s.Pods(tc.Namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}

			for _, s := range objects.Items {
				if s.Name == tc.Name {
					t.Fatal("The object wasn't deleted")
				}
			}
		})
	}
}

func TestClient_Upsert(t *testing.T) {
	fakePod := func(name string, hostname string) *v1.Pod {
		s := &v1.Pod{}
		s.Name = name
		s.Spec.Hostname = hostname
		return s
	}

	cases := map[string]struct {
		Namespace   string
		PreExisting []*v1.Pod
		ToUpsert    *v1.Pod
		Merger      Merger
		ExpectErr   error
		Validate    func(t *testing.T, mockK8s cv1.PodsGetter)
	}{
		"inserts if not found": {
			Namespace: "default",
			ToUpsert:  fakePod("foo", "new"),
			Merger:    nil, // should not be called
			Validate: func(t *testing.T, mockK8s cv1.PodsGetter) {
				obj, err := mockK8s.Pods("default").Get(context.Background(), "foo", metav1.GetOptions{})
				testutil.AssertNil(t, "get err", err)
				testutil.AssertEqual(t, "value", "new", obj.Spec.Hostname)
			},
		},
		"update if found": {
			PreExisting: []*v1.Pod{fakePod("foo", "old")},
			Namespace:   "testing",
			ToUpsert:    fakePod("foo", "new"),
			Merger:      func(n, o *v1.Pod) *v1.Pod { return n },
			Validate: func(t *testing.T, mockK8s cv1.PodsGetter) {
				obj, err := mockK8s.Pods("testing").Get(context.Background(), "foo", metav1.GetOptions{})
				testutil.AssertNil(t, "get err", err)
				testutil.AssertEqual(t, "value", "new", obj.Spec.Hostname)
			},
		},
		"calls merge with right order": {
			Namespace:   "default",
			PreExisting: []*v1.Pod{fakePod("foo", "old")},
			ToUpsert:    fakePod("foo", "new"),
			Merger: func(n, o *v1.Pod) *v1.Pod {
				n.Spec.Hostname = n.Spec.Hostname + "-" + o.Spec.Hostname
				return n
			},
			ExpectErr: nil,
			Validate: func(t *testing.T, mockK8s cv1.PodsGetter) {
				obj, err := mockK8s.Pods("default").Get(context.Background(), "foo", metav1.GetOptions{})
				testutil.AssertNil(t, "get err", err)
				testutil.AssertEqual(t, "value", "new-old", obj.Spec.Hostname)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			mockK8s := testclient.NewSimpleClientset().CoreV1()

			client := NewExampleClient(mockK8s)

			for _, v := range tc.PreExisting {
				_, err := client.Create(context.Background(), tc.Namespace, v)
				testutil.AssertNil(t, "creating preexisting objects", err)
			}

			_, actualErr := client.Upsert(context.Background(), tc.Namespace, tc.ToUpsert, tc.Merger)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			if tc.Validate != nil {
				tc.Validate(t, mockK8s)
			}
		})
	}
}

func TestConditionDeleted(t *testing.T) {
	cases := map[string]struct {
		apiErr error

		wantDone bool
		wantErr  error
	}{
		"not found error": {
			apiErr:   apierrors.NewNotFound(schema.GroupResource{}, "my-obj"),
			wantDone: true,
		},
		"nil error": {
			apiErr:   nil,
			wantDone: false,
		},
		"other error": {
			apiErr:   errors.New("some-error"),
			wantDone: true,
			wantErr:  errors.New("some-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualDone, actualErr := ConditionDeleted(nil, tc.apiErr)

			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
		})
	}
}

func ExampleClient_WaitFor_timeout() {
	pod := &v1.Pod{}
	pod.Name = "object-name"
	mockK8s := testclient.NewSimpleClientset().CoreV1()
	mockK8s.Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	client := NewExampleClient(mockK8s)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	called := 0
	instance, err := client.WaitFor(ctx, "default", "object-name", 100*time.Millisecond, func(_ *v1.Pod) bool {
		called++
		return false
	})
	fmt.Println("Instance:", instance)
	fmt.Println("Error:", err)
	fmt.Println("Called?:", called) // 3 calls, immediately, 100ms later then 100ms after

	// Output: Instance: nil
	// Error: waiting for OperatorConfig timed out
	// Called?: 3
}

func TestWrapPredicate(t *testing.T) {
	cases := map[string]struct {
		apiErr            error
		predicateResponse bool
		conditions        []v1.PodCondition

		wantCall bool
		wantDone bool
		wantErr  error
	}{
		"errors are final": {
			apiErr:   errors.New("some-error"),
			wantDone: true,
			wantErr:  errors.New("some-error"),
		},
		"false conditions results in error": {
			wantDone: true,
			wantErr:  errors.New(`Reason: "SomeReason", Message: "SomeMessage"`),

			// It should be done because of the false conditions, not because
			// the predicate.
			predicateResponse: false,
			conditions: []v1.PodCondition{
				{
					Status:  v1.ConditionFalse,
					Message: "SomeMessage",
					Reason:  "SomeReason",
				},
			},
		},
		"no error returns response (false)": {
			apiErr:            nil,
			predicateResponse: false,
			wantCall:          true,
			wantDone:          false,
		},
		"no error returns response (true)": {
			apiErr:            nil,
			predicateResponse: true,
			wantCall:          true,
			wantDone:          true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			called := false
			wrapped := wrapPredicate(func(_ *v1.Pod) bool {
				called = true
				return tc.predicateResponse
			})
			actualDone, actualErr := wrapped(&v1.Pod{
				Status: v1.PodStatus{
					Conditions: tc.conditions,
				},
			}, tc.apiErr)

			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
			testutil.AssertEqual(t, "predicate called", tc.wantCall, called)
		})
	}
}

func TestCheckConditionTrue(t *testing.T) {
	cases := map[string]struct {
		conditions []v1.PodCondition
		err        error

		wantDone bool
		wantErr  error
	}{
		"error supplied": {
			err: errors.New("test"),

			wantDone: true,
			wantErr:  errors.New("test"),
		},
		"condition missing": {
			conditions: []v1.PodCondition{},
			wantDone:   false,
			wantErr:    nil,
		},
		"condition unknown": {
			conditions: []v1.PodCondition{
				{
					Type:   v1.PodConditionType(ConditionReady),
					Status: v1.ConditionUnknown,
				},
			},
			wantDone: false,
			wantErr:  nil,
		},
		"condition true": {
			conditions: []v1.PodCondition{
				{
					Type:   v1.PodConditionType(ConditionReady),
					Status: v1.ConditionTrue,
				},
			},
			wantDone: true,
			wantErr:  nil,
		},
		"condition false": {
			conditions: []v1.PodCondition{
				{
					Type:    v1.PodConditionType(ConditionReady),
					Status:  v1.ConditionFalse,
					Message: "SomeMessage",
					Reason:  "SomeReason",
				},
			},
			wantDone: true,
			wantErr:  errors.New("checking Ready failed, status: False message: SomeMessage reason: SomeReason"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			pod := &v1.Pod{
				Status: v1.PodStatus{
					Conditions: tc.conditions,
				},
			}

			actualDone, actualErr := checkConditionTrue(pod, tc.err, ConditionReady)

			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
		})
	}
}
