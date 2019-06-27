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

package v1alpha1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	rv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Space) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Space")
}

// ConditionType represents a Service condition value
const (
	// SpaceConditionReady is set when the space is configured
	// and is usable by developers.
	SpaceConditionReady = apis.ConditionReady
	// SpaceConditionNamespaceReady is set when the backing namespace is ready.
	SpaceConditionNamespaceReady apis.ConditionType = "NamespaceReady"
	// SpaceConditionDeveloperRoleReady is set when the developer RBAC role is
	// ready.
	SpaceConditionDeveloperRoleReady apis.ConditionType = "DeveloperRoleReady"
	// SpaceConditionAuditorRoleReady is set when the auditor RBAC role is
	// ready.
	SpaceConditionAuditorRoleReady apis.ConditionType = "AuditorRoleReady"
)

func (status *SpaceStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		SpaceConditionNamespaceReady,
		SpaceConditionDeveloperRoleReady,
		SpaceConditionAuditorRoleReady,
	).Manage(status)
}

// IsReady returns if the space is ready to be used.
func (status *SpaceStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *SpaceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *SpaceStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// MarkNamespaceNotOwned marks the namespace as not being owned by the Space.
func (status *SpaceStatus) MarkNamespaceNotOwned(name string) {
	status.manage().MarkFalse(SpaceConditionNamespaceReady, "NotOwned",
		fmt.Sprintf("There is an existing Namespace %q that we do not own.", name))
}

// MarkDeveloperRoleNotOwned marks the developer role as not being owned by the Space.
func (status *SpaceStatus) MarkDeveloperRoleNotOwned(name string) {
	status.manage().MarkFalse(SpaceConditionDeveloperRoleReady, "NotOwned",
		fmt.Sprintf("There is an existing developer role %q that we do not own.", name))
}

// MarkAuditorRoleNotOwned marks the auditor role as not being owned by the Space.
func (status *SpaceStatus) MarkAuditorRoleNotOwned(name string) {
	status.manage().MarkFalse(SpaceConditionAuditorRoleReady, "NotOwned",
		fmt.Sprintf("There is an existing auditor role %q that we do not own.", name))
}

// PropagateNamespaceStatus copies fields from the Namespace status to Space
// and updates the readiness based on the current phase.
func (status *SpaceStatus) PropagateNamespaceStatus(ns *v1.Namespace) {
	// TODO(josephlewis42): should we copy the namespace's UID into the status?

	switch ns.Status.Phase {
	case v1.NamespaceActive:
		status.manage().MarkTrue(SpaceConditionNamespaceReady)
	case v1.NamespaceTerminating:
		status.manage().MarkFalse(SpaceConditionNamespaceReady, "Terminating", "Namespace is terminating")
	default:
		status.manage().MarkUnknown(SpaceConditionNamespaceReady, "BadPhase", "Namespace entered an unknown phase: %q", ns.Status.Phase)
	}
}

// PropagateDeveloperRoleStatus copies fields from the Role to Space
// and updates the readiness based on the current phase.
func (status *SpaceStatus) PropagateDeveloperRoleStatus(*rv1.Role) {
	// Roles don't have a status field so they just need to exist to be ready.
	status.manage().MarkTrue(SpaceConditionDeveloperRoleReady)
}

// PropagateAuditorRoleStatus copies fields from the Role to Space
// and updates the readiness based on the current phase.
func (status *SpaceStatus) PropagateAuditorRoleStatus(*rv1.Role) {
	// Roles don't have a status field so they just need to exist to be ready.
	status.manage().MarkTrue(SpaceConditionAuditorRoleReady)
}

func (status *SpaceStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
