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

	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *App) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("App")
}

// ConditionType represents a Service condition value
const (
	// AppConditionReady is set when the app is configured
	// and is usable by developers.
	AppConditionReady = apis.ConditionReady
	// AppConditionSourceReady is set when the build is ready.
	AppConditionSourceReady apis.ConditionType = "SourceReady"
	// AppConditionKnativeServiceReady is set when service is ready.
	AppConditionKnativeServiceReady apis.ConditionType = "KnativeServiceReady"
	// AppConditionSpaceReady is used to indicate when the space has an error that
	// causes apps to not reconcile correctly.
	AppConditionSpaceReady apis.ConditionType = "SpaceReady"
)

func (status *AppStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		AppConditionSourceReady,
		AppConditionKnativeServiceReady,
		AppConditionSpaceReady,
	).Manage(status)
}

// GetCondition returns the condition by name.
func (status *AppStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *AppStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// MarkSourceNotOwned marks the Source as not being owned by the App.
func (status *AppStatus) MarkSourceNotOwned(name string) {
	status.manage().MarkFalse(AppConditionSourceReady, "NotOwned",
		fmt.Sprintf("There is an existing Source %q that we do not own.", name))
}

// MarkSourceTemplateError marks the Source as having an error with the template.
func (status *AppStatus) MarkSourceTemplateError(err error) {
	status.manage().MarkFalse(AppConditionSourceReady, "TemplateError",
		fmt.Sprintf("Couldn't populate the Source child template: %s", err))
}

// MarkSourceReconciliationError marks the Source having some error during the
// reconciliation process.
func (status *AppStatus) MarkSourceReconciliationError(context string, err error) {
	status.manage().MarkFalse(AppConditionSourceReady, "ReconciliationError",
		fmt.Sprintf("Error occurred while %s: %s", context, err))
}

// MarkKnativeServiceNotOwned marks the Knative Service as not being owned by
// the App.
func (status *AppStatus) MarkKnativeServiceNotOwned(name string) {
	status.manage().MarkFalse(AppConditionKnativeServiceReady, "NotOwned",
		fmt.Sprintf("There is an existing Knative Service %q that we do not own.", name))
}

// MarkKnativeServiceTemplateError marks the Source as not having an error with
// the template.
func (status *AppStatus) MarkKnativeServiceTemplateError(err error) {
	status.manage().MarkFalse(AppConditionKnativeServiceReady, "TemplateError",
		fmt.Sprintf("Couldn't populate the KnativeService child template: %s", err))
}

// PropagateSourceStatus copies the source status to the app's.
func (status *AppStatus) PropagateSourceStatus(source *Source) {
	status.LatestCreatedSourceName = source.Name

	cond := source.Status.GetCondition(SourceConditionSucceeded)
	switch {
	case cond == nil:
		return
	case cond.IsFalse():
		status.manage().MarkFalse(AppConditionSourceReady, cond.Reason, cond.Message)
	case cond.IsTrue():
		status.LatestReadySourceName = source.Name
		status.SourceStatusFields = source.Status.SourceStatusFields
		status.manage().MarkTrue(AppConditionSourceReady)
	case cond.IsUnknown():
		status.manage().MarkUnknown(AppConditionSourceReady, cond.Reason, cond.Message)
	}
}

// PropagateKnativeServiceStatus updates the Knative service status to reflect
// the underlying service.
func (status *AppStatus) PropagateKnativeServiceStatus(service *serving.Service) {
	cond := service.Status.GetCondition(apis.ConditionReady)
	switch {
	case cond == nil:
		return
	case cond.IsFalse():
		status.manage().MarkFalse(AppConditionKnativeServiceReady, cond.Reason, cond.Message)
	case cond.IsTrue():
		status.ConfigurationStatusFields = service.Status.ConfigurationStatusFields
		status.manage().MarkTrue(AppConditionKnativeServiceReady)
	case cond.IsUnknown():
		status.manage().MarkUnknown(AppConditionKnativeServiceReady, cond.Reason, cond.Message)
	}
}

// MarkSpaceHealthy notes that the space was able to be retrieved and
// defaults can be applied from it.
func (status *AppStatus) MarkSpaceHealthy() {
	status.manage().MarkTrue(AppConditionSpaceReady)
}

// MarkSpaceUnhealthy notes that the space was could not be retrieved.
func (status *AppStatus) MarkSpaceUnhealthy(reason, message string) {
	status.manage().MarkFalse(AppConditionSpaceReady, reason, message)
}
