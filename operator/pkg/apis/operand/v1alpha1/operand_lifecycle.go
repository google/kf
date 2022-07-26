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

// Operand is just used for GC and has no status.
var operandConditions = apis.NewLivingConditionSet(LatestActiveOperandReady, OperandInstalled)

// GetConditionSet retrieves the ConditionSet of Operand. Implements the KRShaped interface.
func (o *Operand) GetConditionSet() apis.ConditionSet {
	return operandConditions
}

// GetStatus retrieves the status of Operand. Implements the KRShaped interface.
func (o *Operand) GetStatus() *duckv1.Status {
	return &o.Status.Status
}

// GetGroupVersionKind returns the GroupVersionKind.
func (o *Operand) GetGroupVersionKind() schema.GroupVersionKind {
	return Kind("Operand")
}

// MarkLatestActiveOperandCreated sets LatestCreatedActiveOperand.
func (os *OperandStatus) MarkLatestActiveOperandCreated(name string) {
	os.LatestCreatedActiveOperand = name
}

// MarkLatestActiveOperandReady sets LatestReadyActiveOperand and
// sets LatestActiveOperandReady condition true.
func (os *OperandStatus) MarkLatestActiveOperandReady(name string) {
	os.LatestReadyActiveOperand = name
	operandConditions.Manage(os).MarkTrue(LatestActiveOperandReady)
}

// ResetLatestCreatedActiveOperand sets LatestCreatedActiveOperand and resets conditions.
func (os *OperandStatus) ResetLatestCreatedActiveOperand(name string) {
	os.LatestCreatedActiveOperand = name
	operandConditions.Manage(os).MarkUnknown(
		LatestActiveOperandReady,
		"Creating",
		"New operand %s being created", name)
	os.MarkOperandInstallNotReady("New operand created. Re-checking health")
}

// MarkOperandInstallFailed sets OperandInstalled condition false.
func (os *OperandStatus) MarkOperandInstallFailed(err error) {
	operandConditions.Manage(os).MarkFalse(
		OperandInstalled,
		"Failed",
		"Failure during operand install: %s", fmt.Sprintf("%+v", err))
}

// MarkOperandInstallNotReady sets OperandInstalled condition false.
func (os *OperandStatus) MarkOperandInstallNotReady(msg string) {
	operandConditions.Manage(os).MarkFalse(OperandInstalled,
		"Waiting",
		"Waiting on Operand. Current state is %s", msg)
}

// MarkOperandPostInstallFailed sets OperandInstalled condition false
// after a post-install failure.
func (os *OperandStatus) MarkOperandPostInstallFailed(err error) {
	operandConditions.Manage(os).MarkFalse(
		OperandInstalled,
		"Failed",
		"Failure during operand post-install: %+v", err)
}

// MarkOperandPostInstallNotReady sets OperandInstalled condition
// false while waiting for post-install.
func (os *OperandStatus) MarkOperandPostInstallNotReady(msg string) {
	operandConditions.Manage(os).MarkFalse(OperandInstalled,
		"Waiting",
		"Waiting on Operand post-install. Current state is %s", msg)
}

// MarkOperandInstallSuccessful sets OperandInstalled condition true.
func (os *OperandStatus) MarkOperandInstallSuccessful() {
	operandConditions.Manage(os).MarkTrue(OperandInstalled)
}

// IsReady returns if the status is ready.
func (os *OperandStatus) IsReady() bool {
	return operandConditions.Manage(os).IsHappy()
}

// IsFalse returns true if the TopLevelCondition is false and it's not waiting.
func (os *OperandStatus) IsFalse() bool {
	c := operandConditions.Manage(os).GetTopLevelCondition()
	return c.IsFalse() && c.Reason != "Waiting"
}

// InitializeConditions sets the initial values to the operandConditions.
func (os *OperandStatus) InitializeConditions() {
	operandConditions.Manage(os).InitializeConditions()
}
