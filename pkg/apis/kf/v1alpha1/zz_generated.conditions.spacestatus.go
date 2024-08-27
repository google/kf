// Copyright 2024 Google LLC
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

// This file was generated with conditiongen/generator.go, DO NOT EDIT IT.

package v1alpha1

import (
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// ConditionType represents a Service condition value
const (

	// SpaceConditionReady is set when the CRD is configured and is usable.
	SpaceConditionReady = apis.ConditionReady

	// SpaceConditionNamespaceReady is set when the child
	// resource(s) Namespace is/are ready.
	SpaceConditionNamespaceReady apis.ConditionType = "NamespaceReady"

	// SpaceConditionBuildServiceAccountReady is set when the child
	// resource(s) BuildServiceAccount is/are ready.
	SpaceConditionBuildServiceAccountReady apis.ConditionType = "BuildServiceAccountReady"

	// SpaceConditionBuildSecretReady is set when the child
	// resource(s) BuildSecret is/are ready.
	SpaceConditionBuildSecretReady apis.ConditionType = "BuildSecretReady"

	// SpaceConditionBuildRoleReady is set when the child
	// resource(s) BuildRole is/are ready.
	SpaceConditionBuildRoleReady apis.ConditionType = "BuildRoleReady"

	// SpaceConditionBuildRoleBindingReady is set when the child
	// resource(s) BuildRoleBinding is/are ready.
	SpaceConditionBuildRoleBindingReady apis.ConditionType = "BuildRoleBindingReady"

	// SpaceConditionIngressGatewayReady is set when the child
	// resource(s) IngressGateway is/are ready.
	SpaceConditionIngressGatewayReady apis.ConditionType = "IngressGatewayReady"

	// SpaceConditionRuntimeConfigReady is set when the child
	// resource(s) RuntimeConfig is/are ready.
	SpaceConditionRuntimeConfigReady apis.ConditionType = "RuntimeConfigReady"

	// SpaceConditionNetworkConfigReady is set when the child
	// resource(s) NetworkConfig is/are ready.
	SpaceConditionNetworkConfigReady apis.ConditionType = "NetworkConfigReady"

	// SpaceConditionBuildConfigReady is set when the child
	// resource(s) BuildConfig is/are ready.
	SpaceConditionBuildConfigReady apis.ConditionType = "BuildConfigReady"

	// SpaceConditionBuildNetworkPolicyReady is set when the child
	// resource(s) BuildNetworkPolicy is/are ready.
	SpaceConditionBuildNetworkPolicyReady apis.ConditionType = "BuildNetworkPolicyReady"

	// SpaceConditionAppNetworkPolicyReady is set when the child
	// resource(s) AppNetworkPolicy is/are ready.
	SpaceConditionAppNetworkPolicyReady apis.ConditionType = "AppNetworkPolicyReady"

	// SpaceConditionRoleBindingsReady is set when the child
	// resource(s) RoleBindings is/are ready.
	SpaceConditionRoleBindingsReady apis.ConditionType = "RoleBindingsReady"

	// SpaceConditionClusterRoleReady is set when the child
	// resource(s) ClusterRole is/are ready.
	SpaceConditionClusterRoleReady apis.ConditionType = "ClusterRoleReady"

	// SpaceConditionClusterRoleBindingsReady is set when the child
	// resource(s) ClusterRoleBindings is/are ready.
	SpaceConditionClusterRoleBindingsReady apis.ConditionType = "ClusterRoleBindingsReady"

	// SpaceConditionIAMPolicyReady is set when the child
	// resource(s) IAMPolicy is/are ready.
	SpaceConditionIAMPolicyReady apis.ConditionType = "IAMPolicyReady"
)

func (status *SpaceStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		SpaceConditionNamespaceReady,
		SpaceConditionBuildServiceAccountReady,
		SpaceConditionBuildSecretReady,
		SpaceConditionBuildRoleReady,
		SpaceConditionBuildRoleBindingReady,
		SpaceConditionIngressGatewayReady,
		SpaceConditionRuntimeConfigReady,
		SpaceConditionNetworkConfigReady,
		SpaceConditionBuildConfigReady,
		SpaceConditionBuildNetworkPolicyReady,
		SpaceConditionAppNetworkPolicyReady,
		SpaceConditionRoleBindingsReady,
		SpaceConditionClusterRoleReady,
		SpaceConditionClusterRoleBindingsReady,
		SpaceConditionIAMPolicyReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *SpaceStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *SpaceStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(SpaceConditionReady, "Terminating", "resource is terminating")
}

// GetCondition returns the condition by name.
func (status *SpaceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *SpaceStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// NamespaceCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) NamespaceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionNamespaceReady, "Namespace")
}

// BuildServiceAccountCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildServiceAccountCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildServiceAccountReady, "BuildServiceAccount")
}

// BuildSecretCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildSecretReady, "BuildSecret")
}

// BuildRoleCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildRoleCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildRoleReady, "BuildRole")
}

// BuildRoleBindingCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildRoleBindingCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildRoleBindingReady, "BuildRoleBinding")
}

// IngressGatewayCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) IngressGatewayCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionIngressGatewayReady, "IngressGateway")
}

// RuntimeConfigCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) RuntimeConfigCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionRuntimeConfigReady, "RuntimeConfig")
}

// NetworkConfigCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) NetworkConfigCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionNetworkConfigReady, "NetworkConfig")
}

// BuildConfigCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildConfigCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildConfigReady, "BuildConfig")
}

// BuildNetworkPolicyCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) BuildNetworkPolicyCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionBuildNetworkPolicyReady, "BuildNetworkPolicy")
}

// AppNetworkPolicyCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) AppNetworkPolicyCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionAppNetworkPolicyReady, "AppNetworkPolicy")
}

// RoleBindingsCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) RoleBindingsCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionRoleBindingsReady, "RoleBindings")
}

// ClusterRoleCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) ClusterRoleCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionClusterRoleReady, "ClusterRole")
}

// ClusterRoleBindingsCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) ClusterRoleBindingsCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionClusterRoleBindingsReady, "ClusterRoleBindings")
}

// IAMPolicyCondition gets a manager for the state of the child resource.
func (status *SpaceStatus) IAMPolicyCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), SpaceConditionIAMPolicyReady, "IAMPolicy")
}

func (status *SpaceStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
