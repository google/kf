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

// ClusterActiveOperand is just used for GC and has no status.
var clusterActiveOperandConditions = apis.NewLivingConditionSet(OwnerRefsInjected, NamespaceDelegatesReady)

// GetConditionSet retrieves the ConditionSet of ClusterActiveOperand. Implements the KRShaped interface.
func (ao *ClusterActiveOperand) GetConditionSet() apis.ConditionSet {
	return clusterActiveOperandConditions
}

// GetStatus implements the duckv1.Status interface.
func (ao *ClusterActiveOperand) GetStatus() *duckv1.Status {
	return &ao.Status.Status
}

// GetGroupVersionKind returns the GroupVersionKind.
func (*ClusterActiveOperand) GetGroupVersionKind() schema.GroupVersionKind {
	return Kind("ClusterActiveOperand")
}

// MarkOwnerRefsInjected shows that the live references have had an
// owner reference injected succesfully.
func (aos *ClusterActiveOperandStatus) MarkOwnerRefsInjected() {
	clusterActiveOperandConditions.Manage(aos).MarkTrue(OwnerRefsInjected)
}

// SetClusterLive sets ClusterLive.
func (aos *ClusterActiveOperandStatus) SetClusterLive(live ...LiveRef) {
	aos.ClusterLive = live
}

// SetDelegates sets Delegates.
func (aos *ClusterActiveOperandStatus) SetDelegates(delegate ...DelegateRef) {
	aos.Delegates = delegate
}

// MarkOwnerRefsInjectedFailed shows that the live references failed to have
// an owner reference injected, this may be transient or permanent.
func (aos *ClusterActiveOperandStatus) MarkOwnerRefsInjectedFailed(msg string) {
	clusterActiveOperandConditions.Manage(aos).MarkFalse(OwnerRefsInjected, "Error", fmt.Sprintf("Failed to inject ownerrefs: %s", msg))
}

// MarkNamespaceDelegatesReady shows that the namespace delegates are all ready.
func (aos *ClusterActiveOperandStatus) MarkNamespaceDelegatesReady() {
	clusterActiveOperandConditions.Manage(aos).MarkTrue(NamespaceDelegatesReady)
}

// IsNamespaceDelegatesReady returns whether condition NamespaceDelegatesReady is true.
func (aos *ClusterActiveOperandStatus) IsNamespaceDelegatesReady() bool {
	return clusterActiveOperandConditions.Manage(aos).GetCondition(NamespaceDelegatesReady).IsTrue()
}

// MarkNamespaceDelegatesReadyFailed surfaces errors from delegates that have failed.
func (aos *ClusterActiveOperandStatus) MarkNamespaceDelegatesReadyFailed(msg string) {
	clusterActiveOperandConditions.Manage(aos).MarkFalse(NamespaceDelegatesReady, "Error", fmt.Sprintf("Failed to add ownerrefs to namespaces: %s", msg))
}

// IsReady returns if the status is ready.
func (aos *ClusterActiveOperandStatus) IsReady() bool {
	return clusterActiveOperandConditions.Manage(aos).IsHappy()
}

// InitializeConditions sets the initial values to the clusterActiveOperandConditions.
func (aos *ClusterActiveOperandStatus) InitializeConditions() {
	clusterActiveOperandConditions.Manage(aos).InitializeConditions()
}
