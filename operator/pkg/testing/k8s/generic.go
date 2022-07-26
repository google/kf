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

package k8s

import (
	"testing"

	otesting "kf-operator/pkg/testing"
	mftest "kf-operator/pkg/testing/manifestival"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	rtesting "knative.dev/pkg/reconciler/testing"
)

var (
	// OperatorScheme is Scheme for operator.
	OperatorScheme = otesting.NewScheme()
)

// AddAnnotation updates annotations in Object.
// TODO(b/160998627): This is inconsistent with other mutators which are specific
// to the k8s object being modified. If we start to codegen these k8s helper files
// we should collapse this into a per-resource type method.
func AddAnnotation(obj mftest.Object, annotation string, value string) mftest.Object {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		obj.SetAnnotations(make(map[string]string))
	}
	obj.GetAnnotations()[annotation] = value
	return obj
}

// AddLabel updates labels in Object.
// TODO(b/160998627): This is inconsistent with other mutators which are specific
// to the k8s object being modified. If we start to codegen these k8s helper files
// we should collapse this into a per-resource type method.
func AddLabel(obj mftest.Object, label string, value string) mftest.Object {
	labels := obj.GetLabels()
	if labels == nil {
		obj.SetLabels(make(map[string]string))
	}
	obj.GetLabels()[label] = value
	return obj
}

// SetOwnerReferences sets OwnerReferences in Object.
func SetOwnerReferences(obj mftest.Object, o ...metav1.OwnerReference) mftest.Object {
	obj.SetOwnerReferences(o)
	return obj
}

// ToRuntimeObjs converts mftest.Objects to runtime.Objects.
func ToRuntimeObjs(in []mftest.Object) []runtime.Object {
	asRuntimeObj := make([]runtime.Object, len(in))
	for i, obj := range in {
		asRuntimeObj[i] = obj
	}
	return asRuntimeObj
}

// CheckCreates checks CreateActions.
func CheckCreates(t *testing.T, actions []ktesting.CreateAction, desiredObjects ...mftest.Object) {
	wants := ToRuntimeObjs(desiredObjects)
	if len(wants) != len(actions) {
		t.Fatalf("Wanted %+v actions, got [%+v]", len(wants), actions)
	}
	for i := 0; i < len(wants); i++ {
		want, got := &unstructured.Unstructured{}, &unstructured.Unstructured{}
		OperatorScheme.Convert(wants[i], want, nil)
		OperatorScheme.Convert(actions[i].GetObject(), got, nil)
		// Ignore empty/nil map/list entries.
		if diff := cmp.Diff(want, got, append(mftest.CommonOptions, mftest.IgnoreLastAppliedConfig)...); diff != "" {
			t.Fatalf("Unexpected create for step %d (-want, +got): %s", i, diff)
		}
	}
}

// CheckUpdates checks UpdateActions.
func CheckUpdates(t *testing.T, actions []ktesting.UpdateAction, desiredObjects ...mftest.Object) {
	wants := ToRuntimeObjs(desiredObjects)
	if len(wants) != len(actions) {
		t.Fatalf("Wanted %+v actions, got [%+v]", len(wants), actions)
	}
	for i := 0; i < len(wants); i++ {
		want, got := &unstructured.Unstructured{}, &unstructured.Unstructured{}
		OperatorScheme.Convert(wants[i], want, nil)
		OperatorScheme.Convert(actions[i].GetObject(), got, nil)
		// Ignore empty/nil map/list entries.
		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(),
			cmpopts.IgnoreMapEntries(func(k, v interface{}) bool {
				if v == nil {
					return true
				}
				if list, converted := v.([]interface{}); converted {
					return len(list) == 0
				}
				if dict, converted := v.(map[string]interface{}); converted {
					return len(dict) == 0
				}
				return false
			}), mftest.IgnoreLastAppliedConfig); diff != "" {
			t.Fatalf("Unexpected create for step %d (-want, +got): %s", i, diff)
		}
	}
}

// CheckNoMutates checks there are no mutating actions.
func CheckNoMutates(t *testing.T, actionList *rtesting.ActionRecorderList) {
	actions, err := actionList.ActionsByVerb()
	if err != nil {
		t.Fatalf("Error sorting actions %+v.", err)
	}
	CheckNoMutatesFromActions(t, actions)
}

// CheckNoMutatesFromActions checks there are no mutating actions.
func CheckNoMutatesFromActions(t *testing.T, actions rtesting.Actions) {
	if len(actions.Creates)+len(actions.Deletes)+len(actions.Updates)+len(actions.DeleteCollections)+len(actions.Patches) > 0 {
		t.Fatalf("Wanted no mutating actions. Got [%+v]", actions)
	}
}

// ValidateNonExpectedActions validates that there are no creates, deletes, updates,
// deletecollections or patches actions.
func ValidateNonExpectedActions(t *testing.T, actions rtesting.Actions, expectedActions ...string) {
	if !actionInSlice(expectedActions, "creates") && len(actions.Creates) > 0 {
		t.Fatalf("Saw creates %+v", actions.Creates)
	}
	if !actionInSlice(expectedActions, "deletes") && len(actions.Deletes) > 0 {
		t.Fatalf("Saw deletes %+v", actions.Deletes)
	}
	if !actionInSlice(expectedActions, "updates") && len(actions.Updates) > 0 {
		t.Fatalf("Saw updates %+v", actions.Updates)
	}
	if !actionInSlice(expectedActions, "deletecollections") && len(actions.DeleteCollections) > 0 {
		t.Fatalf("Saw deletecolections %+v", actions.DeleteCollections)
	}
	if !actionInSlice(expectedActions, "patches") && len(actions.Patches) > 0 {
		t.Fatalf("Saw patches %v", actions.Patches)
	}
}

func actionInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
