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

	"k8s.io/apimachinery/pkg/runtime/schema"
	apistest "knative.dev/pkg/apis/testing"
)

func TestGetGroupVersionKind(t *testing.T) {
	cr := &ActiveOperand{}
	want := schema.GroupVersionKind{
		Group:   GroupName,
		Version: SchemaVersion,
		Kind:    "ActiveOperand",
	}
	if got := cr.GetGroupVersionKind(); got != want {
		t.Errorf("GroupVersionKind: got %v, want: %v", got, want)
	}
}

func TestReady(t *testing.T) {
	ao := &ActiveOperand{}
	ao.Status.MarkOwnerRefsInjected()

	if !ao.Status.IsReady() {
		t.Errorf("%+v should be happy, but is not.", ao)
	}
}

func TestMarkOwnerRefsInjectedFailed(t *testing.T) {
	ao := &ActiveOperand{}
	ao.Status.MarkOwnerRefsInjectedFailed("test")
	apistest.CheckConditionFailed(&ao.Status, OwnerRefsInjected, t)
	e := ao.Status.GetCondition(OwnerRefsInjected)
	if got, want := e.Reason, "Error"; got != want {
		t.Errorf("Condition Reason: got %v, want %v", got, want)
	}

	if ao.Status.IsReady() {
		t.Errorf("%+v should NOT be happy, but is.", ao)
	}
}
