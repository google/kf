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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var conditions = apis.NewLivingConditionSet(
	KfInstallSucceeded,
)

// InitializeConditions sets the initial values to the kfConditions.
func (kfs *KfSystemStatus) InitializeConditions() {
	conditions.Manage(kfs).InitializeConditions()
}

// GetConditionSet retrieves the ConditionSet of Kf. Implements the KRShaped interface.
func (*KfSystem) GetConditionSet() apis.ConditionSet {
	return conditions
}

// GetStatus retrieves the status of Kf. Implements the KRShaped interface.
func (kf *KfSystem) GetStatus() *duckv1.Status {
	return &kf.Status.Status
}

// GetGroupVersionKind returns the GroupVersionKind.
func (*KfSystem) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersionKind
}

// IsEnabled returns if the enabled field is set to true.
func (spec *KfSpec) IsEnabled() bool {
	return spec.Enabled == nil || *spec.Enabled
}

// IsReady returns if the status is ready.
func (kfs *KfSystemStatus) IsReady() bool {
	return conditions.Manage(kfs).IsHappy()
}

// IsKfInstallFailed returns whether Kf failed to be installed.
func (kfs *KfSystemStatus) IsKfInstallFailed() bool {
	return kfs.GetCondition(KfInstallSucceeded).IsFalse()
}

// MarkKfInstallFailed marks Kf install failed.
func (kfs *KfSystemStatus) MarkKfInstallFailed(msg string) {
	conditions.Manage(kfs).MarkFalse(
		KfInstallSucceeded,
		"Error",
		"Kf Install failed with message: %s", msg)
}

// MarkKfInstallNotReady marks Kf install not ready.
func (kfs *KfSystemStatus) MarkKfInstallNotReady() {
	conditions.Manage(kfs).MarkUnknown(
		KfInstallSucceeded,
		"NotReady",
		"Kf Install waiting on deployments")
}

// MarkKfInstallSucceeded marks Kf install succeeded.
func (kfs *KfSystemStatus) MarkKfInstallSucceeded(version string) {
	conditions.Manage(kfs).MarkTrue(KfInstallSucceeded)
	kfs.KfVersion = version
}
