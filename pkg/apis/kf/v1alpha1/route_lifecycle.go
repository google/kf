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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Route) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Route")
}

// ConditionType represents a Service condition value
const (
	// RouteConditionReady is set when the Route is configured
	// and is usable by developers.
	RouteConditionReady = apis.ConditionReady
	// RouteConditionVirtualServiceReady is set when the backing
	// VirtualService is ready.
	RouteConditionVirtualServiceReady apis.ConditionType = "VirtualServiceReady"
)

func (status *RouteStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		RouteConditionVirtualServiceReady,
	).Manage(status)
}

// IsReady returns if the Route is ready to be used.
func (status *RouteStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *RouteStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *RouteStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// MarkVirtualServiceNotOwned marks the VirtualService as not being owned by
// the Route.
func (status *RouteStatus) MarkVirtualServiceNotOwned(name string) {
	status.manage().MarkFalse(RouteConditionVirtualServiceReady, "NotOwned",
		fmt.Sprintf("There is an existing VirtualService %q that we do not own.", name))
}

// PropagateVirtualServiceStatus copies fields from the VirtualService status to
// Route and updates the readiness based on the current phase.
func (status *RouteStatus) PropagateVirtualServiceStatus(vs *networking.VirtualService) {
	// VirtualService don't have a status field so they just need to exist to
	// be ready.
	status.manage().MarkTrue(RouteConditionVirtualServiceReady)
}

func (status *RouteStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
