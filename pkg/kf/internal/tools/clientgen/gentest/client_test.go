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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	testclient "k8s.io/client-go/kubernetes/fake"
	cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func ExampleList_Filter() {
	first := v1.Pod{}
	first.Name = "ok"

	second := v1.Pod{}
	second.Name = "name-too-long-to-pass"

	list := List{first, second}

	filtered := list.Filter(func(s *v1.Pod) bool {
		return len(s.Name) < 8
	})

	fmt.Println("Results")
	for _, v := range filtered {
		fmt.Println("-", v.Name)
	}

	// Output: Results
	// - ok
}

func ExampleLabelSetMutator() {
	out := &v1.Pod{}
	managedAdder := LabelSetMutator(map[string]string{"managed-by": "kf"})

	managedAdder(out)
	fmt.Printf("Labels: %v", out.Labels)

	// Output: Labels: map[managed-by:kf]
}

func TestClient_invariant(t *testing.T) {
	// This test validates that the mutators are applied to read and
	// write operations.
	mockK8s := testclient.NewSimpleClientset().CoreV1()

	client := NewExampleClient(mockK8s)

	t.Run("create", func(t *testing.T) {
		good := &v1.Pod{}
		good.Name = "created-through-client"

		if _, err := client.Create("default", good); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("list", func(t *testing.T) {
		out, err := client.List("default")
		testutil.AssertNil(t, "list err", err)

		testutil.AssertEqual(t, "object count", 1, len(out))
	})

	t.Run("transform", func(t *testing.T) {
		transformResult, err := client.Transform("default", "created-through-client", func(s *v1.Pod) error {
			s.Labels["mutated"] = "true"

			return nil
		})
		testutil.AssertNil(t, "transform err", err)

		modified, err := client.Get("default", "created-through-client")
		testutil.AssertNil(t, "get err", err)

		testutil.AssertEqual(t, "mutated label", "true", modified.Labels["mutated"])
		testutil.AssertEqual(t, "transformResult", transformResult, modified)
	})
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Namespace string
		Name      string
		Options   []DeleteOption
		ExpectErr error
		Setup     func(mockK8s cv1.PodsGetter)
	}{
		"object does not exist": {
			Namespace: "default",
			Name:      "some-object",
			Options:   []DeleteOption{},
			ExpectErr: errors.New(`couldn't delete the OperatorConfig with the name "some-object": pods "some-object" not found`),
		},
		"object exists": {
			Namespace: "my-namespace",
			Name:      "some-object",
			Options:   []DeleteOption{},
			Setup: func(mockK8s cv1.PodsGetter) {
				obj := &v1.Pod{}
				obj.Name = "some-object"
				mockK8s.Pods("my-namespace").Create(obj)
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

			actualErr := client.Delete(tc.Namespace, tc.Name, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			objects, err := mockK8s.Pods(tc.Namespace).List(metav1.ListOptions{})
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
				obj, err := mockK8s.Pods("default").Get("foo", metav1.GetOptions{})
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
				obj, err := mockK8s.Pods("testing").Get("foo", metav1.GetOptions{})
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
				obj, err := mockK8s.Pods("default").Get("foo", metav1.GetOptions{})
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
				_, err := client.Create(tc.Namespace, v)
				testutil.AssertNil(t, "creating preexisting objects", err)
			}

			_, actualErr := client.Upsert(tc.Namespace, tc.ToUpsert, tc.Merger)
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

func ExampleDiffWrapper_noDiff() {
	obj := &v1.Pod{}

	wrapper := DiffWrapper(os.Stdout, func(s *v1.Pod) error {
		// don't mutate the object
		return nil
	})

	wrapper(obj)

	// Output: No changes
}

func ExampleDiffWrapper_changes() {
	obj := &v1.Pod{}
	obj.Spec.Hostname = "opaque"

	contents := &bytes.Buffer{}
	wrapper := DiffWrapper(contents, func(obj *v1.Pod) error {
		obj.Spec.Hostname = "docker-creds"
		return nil
	})

	fmt.Println("Error:", wrapper(obj))
	firstLine := strings.Split(contents.String(), "\n")[0]
	fmt.Println("First line:", firstLine)

	// Output: Error: <nil>
	// First line: OperatorConfig Diff (-old +new):
}

func ExampleDiffWrapper_err() {
	obj := &v1.Pod{}

	wrapper := DiffWrapper(os.Stdout, func(_ *v1.Pod) error {
		return errors.New("some-error")
	})

	fmt.Println(wrapper(obj))

	// Output: some-error
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

func ExampleClient_WaitForE_conditionDeleted() {
	mockK8s := testclient.NewSimpleClientset().CoreV1()
	client := NewExampleClient(mockK8s)

	instance, err := client.WaitForE(context.Background(), "default", "obj-name", 1*time.Second, ConditionDeleted)
	fmt.Println("Instance:", instance)
	fmt.Println("Error:", err)

	// Output: Instance: nil
	// Error: <nil>
}

func ExampleClient_WaitForE_timeout() {
	mockK8s := testclient.NewSimpleClientset().CoreV1()
	client := NewExampleClient(mockK8s)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	called := 0
	instance, err := client.WaitForE(ctx, "default", "object-name", 100*time.Millisecond, func(_ *v1.Pod, _ error) (bool, error) {
		called++
		return false, nil
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

		wantCall bool
		wantDone bool
		wantErr  error
	}{
		"errors are final": {
			apiErr:   errors.New("some-error"),
			wantDone: true,
			wantErr:  errors.New("some-error"),
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
			actualDone, actualErr := wrapped(nil, tc.apiErr)

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
				v1.PodCondition{
					Type:   v1.PodConditionType(ConditionReady),
					Status: v1.ConditionUnknown,
				},
			},
			wantDone: false,
			wantErr:  nil,
		},
		"condition true": {
			conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodConditionType(ConditionReady),
					Status: v1.ConditionTrue,
				},
			},
			wantDone: true,
			wantErr:  nil,
		},
		"condition false": {
			conditions: []v1.PodCondition{
				v1.PodCondition{
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
