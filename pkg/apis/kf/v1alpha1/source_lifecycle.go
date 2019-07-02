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

	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/pkg/apis"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Source) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Source")
}

const (
	// SourceConditionReady is set when the space is configured
	// and is usable by developers.
	SourceConditionReady = apis.ConditionReady
	// SourceConditionBuildReady is set when the backing Build is ready.
	SourceConditionBuildReady apis.ConditionType = "BuildReady"
)

func (status *SourceStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		SourceConditionBuildReady,
	).Manage(status)
}

// IsReady returns if the space is ready to be used.
func (status *SourceStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *SourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *SourceStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// MarkBuildNotOwned marks the Build as not being owned by the Source.
func (status *SourceStatus) MarkBuildNotOwned(name string) {
	status.manage().MarkFalse(SourceConditionBuildReady, "NotOwned",
		fmt.Sprintf("There is an existing Build %q that we do not own.", name))
}

// PropagateBuildStatus copies fields from the Build status to Space
// and updates the readiness based on the current phase.
func (status *SourceStatus) PropagateBuildStatus(build *v1alpha1.Build) {

	if build == nil {
		return
	}

	for _, condition := range build.Status.GetConditions() {

		c := apis.Condition{
			Type:               apis.ConditionType(string(condition.Type)),
			Status:             condition.Status,
			Severity:           apis.ConditionSeverity(string(condition.Severity)),
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		}

		status.manage().SetCondition(c)

		if condition.Type == "Succeeded" && condition.Status == "True" {
			status.Image = build.Spec.Template.Arguments[0].Value
		}
	}
}

func (status *SourceStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
