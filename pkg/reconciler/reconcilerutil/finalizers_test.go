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

package reconcilerutil

import (
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHasFinalizer(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		obj       metav1.Object
		finalizer string
		want      bool
	}{
		"not included": {
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{"finalizer-a"},
				},
			},
			finalizer: "finalizer-b",
			want:      false,
		},
		"included": {
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{"finalizer-a"},
				},
			},
			finalizer: "finalizer-a",
			want:      true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := HasFinalizer(tc.obj, tc.finalizer)

			testutil.AssertEqual(t, "finalizer", tc.want, got)
		})
	}
}

func ExampleAddFinalizer() {
	const finalizerName = "example-finalizer"
	original := &corev1.Pod{}

	fmt.Println("original HasFinalizer:", HasFinalizer(original, finalizerName))

	// Validate that the original has not been scheduled for deletion.
	if original.GetDeletionTimestamp().IsZero() {
		// If the object doesn't have a finalizer, add it:
		if !HasFinalizer(original, finalizerName) {
			// Don't modify the informer's copy
			toUpdate := original.DeepCopy()
			AddFinalizer(toUpdate, finalizerName)

			// XXX: Call Update() on the object here to set the finalizer
			fmt.Println("toUpdate HasFinalizer:", HasFinalizer(toUpdate, finalizerName))
		}
	}
	fmt.Println("original HasFinalizer after:", HasFinalizer(original, finalizerName))

	// Output: original HasFinalizer: false
	// toUpdate HasFinalizer: true
	// original HasFinalizer after: false
}

func ExampleRemoveFinalizer() {
	const finalizerName = "example-finalizer"

	// Set up an object that's pending deletion with a finalizer.
	original := &corev1.Pod{}
	original.SetDeletionTimestamp(&metav1.Time{})
	original.Finalizers = []string{finalizerName}

	fmt.Println("original HasFinalizer:", HasFinalizer(original, finalizerName))

	// First validate that the original has been scheduled for deletion.
	if original.GetDeletionTimestamp() != nil {
		if HasFinalizer(original, finalizerName) {
			// Don't modify the informer's copy
			toUpdate := original.DeepCopy()
			// XXX: Do cleanup, if successful remove the finalizer
			// otherwise if using a Kf object, call PropagateDeletionBlockedStatus()
			// to tell the user the deletion is currently blocked.
			RemoveFinalizer(toUpdate, finalizerName)

			// XXX: Call Update() on the object here to set the finalizer
			fmt.Println("toUpdate HasFinalizer:", HasFinalizer(toUpdate, finalizerName))
		}
	}

	fmt.Println("original HasFinalizer after:", HasFinalizer(original, finalizerName))

	// Output: original HasFinalizer: true
	// toUpdate HasFinalizer: false
	// original HasFinalizer after: true
}
