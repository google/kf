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

func TestClusterGetGroupVersionKind(t *testing.T) {
	cr := &ClusterActiveOperand{}
	want := schema.GroupVersionKind{
		Group:   GroupName,
		Version: SchemaVersion,
		Kind:    "ClusterActiveOperand",
	}
	if got := cr.GetGroupVersionKind(); got != want {
		t.Errorf("GroupVersionKind: got %v, want: %v", got, want)
	}
}

func TestClusterReady(t *testing.T) {
	ao := &ClusterActiveOperand{}
	ao.Status.MarkOwnerRefsInjected()
	ao.Status.MarkNamespaceDelegatesReady()

	if !ao.Status.IsReady() {
		t.Errorf("%+v should be happy, but is not.", ao)
	}
}

func TestClusterMarkOwnerRefsInjectedFailed(t *testing.T) {
	ao := &ClusterActiveOperand{}
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

func TestClusterMarkNamespaceDelegatesReadyFailed(t *testing.T) {
	ao := &ClusterActiveOperand{}
	ao.Status.MarkNamespaceDelegatesReadyFailed("test")
	apistest.CheckConditionFailed(&ao.Status, NamespaceDelegatesReady, t)
	e := ao.Status.GetCondition(NamespaceDelegatesReady)
	if got, want := e.Reason, "Error"; got != want {
		t.Errorf("Condition Reason: got %v, want %v", got, want)
	}

	if ao.Status.IsReady() {
		t.Errorf("%+v should NOT be happy, but is.", ao)
	}
}
