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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// ActiveOperand is just used for GC and has no status.
var activeOperandConditions = apis.NewLivingConditionSet(OwnerRefsInjected)

// GetConditionSet retrieves the ConditionSet of ActiveOperand. Implements the KRShaped interface.
func (ao *ActiveOperand) GetConditionSet() apis.ConditionSet {
	return activeOperandConditions
}

// GetStatus implements the duckv1.Status interface.
func (ao *ActiveOperand) GetStatus() *duckv1.Status {
	return &ao.Status.Status
}

// GetGroupVersionKind returns the GroupVersionKind.
func (*ActiveOperand) GetGroupVersionKind() schema.GroupVersionKind {
	return Kind("ActiveOperand")
}

// MarkOwnerRefsInjected shows that the live references have had an
// owner reference injected succesfully.
func (aos *ActiveOperandStatus) MarkOwnerRefsInjected() {
	activeOperandConditions.Manage(aos).MarkTrue(OwnerRefsInjected)
}

// MarkOwnerRefsInjectedFailed shows that the live references failed to have
// an owner reference injected, this may be transient or permanent.
func (aos *ActiveOperandStatus) MarkOwnerRefsInjectedFailed(msg string) {
	activeOperandConditions.Manage(aos).MarkFalse(OwnerRefsInjected, "Error", fmt.Sprintf("Failed to inject ownerrefs: %s", msg))
}

// IsReady returns if the status is ready.
func (aos *ActiveOperandStatus) IsReady() bool {
	return activeOperandConditions.Manage(aos).IsHappy()
}

// InitializeConditions sets the initial values to the activeOperandConditions.
func (aos *ActiveOperandStatus) InitializeConditions() {
	activeOperandConditions.Manage(aos).InitializeConditions()
}
