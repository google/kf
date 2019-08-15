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
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/api/core/v1"
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
	// AppConditionRouteReady is set when route is ready.
	AppConditionRouteReady apis.ConditionType = "RouteReady"
	// AppConditionEnvVarSecretReady is set when env var secret is ready.
	AppConditionEnvVarSecretReady apis.ConditionType = "EnvVarSecretReady"
	// AppConditionServiceBindingsReady is set when all service bindings are ready.
	AppConditionServiceBindingsReady apis.ConditionType = "ServiceBindingsReady"
)

func (status *AppStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		AppConditionSourceReady,
		AppConditionKnativeServiceReady,
		AppConditionSpaceReady,
		AppConditionEnvVarSecretReady,
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

// SourceCondition gets a manager for the state of the source.
func (status *AppStatus) SourceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionSourceReady, "Source")
}

// KnativeServiceCondition gets a manager for the state of the Knative Service.
func (status *AppStatus) KnativeServiceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionKnativeServiceReady, "Knative Service")
}

// RouteCondition gets a manager for the state of the kf Route.
func (status *AppStatus) RouteCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionRouteReady, "Route")
}

// EnvVarSecretCondition gets a manager for the state of the env var secret.
func (status *AppStatus) EnvVarSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionEnvVarSecretReady, "Env Var Secret")
}

// ServiceBindingCondition gets a manager for the state of the service bindings.
func (status *AppStatus) ServiceBindingCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), AppConditionServiceBindingsReady, "Service Bindings")
}

// PropagateSourceStatus copies the source status to the app's.
func (status *AppStatus) PropagateSourceStatus(source *Source) {
	status.LatestCreatedSourceName = source.Name

	cond := source.Status.GetCondition(SourceConditionSucceeded)
	if PropagateCondition(status.manage(), AppConditionSourceReady, cond) {
		status.LatestReadySourceName = source.Name
		status.SourceStatusFields = source.Status.SourceStatusFields
	}
}

// PropagateKnativeServiceStatus updates the Knative service status to reflect
// the underlying service.
func (status *AppStatus) PropagateKnativeServiceStatus(service *serving.Service) {
	cond := service.Status.GetCondition(apis.ConditionReady)

	if PropagateCondition(status.manage(), AppConditionKnativeServiceReady, cond) {
		status.ConfigurationStatusFields = service.Status.ConfigurationStatusFields
		status.RouteStatusFields = service.Status.RouteStatusFields
	}
}

// PropagateEnvVarSecretStatus updates the env var secret readiness status.
func (status *AppStatus) PropagateEnvVarSecretStatus(secret *v1.Secret) {
	status.manage().MarkTrue(AppConditionEnvVarSecretReady)
}

// PropagateServiceBindingsStatus updates the service binding readiness status.
func (status *AppStatus) PropagateServiceBindingsStatus(bindings []servicecatalogv1beta1.ServiceBinding) {

	var bindingNames []string
	for _, binding := range bindings {
		bindingNames = append(bindingNames, binding.Name)
	}
	status.ServiceBindingNames = bindingNames

	for _, binding := range bindings {

		for _, cond := range binding.Status.Conditions {
			if cond.Type != servicecatalogv1beta1.ServiceBindingConditionReady {
				continue
			}
			if cond.Status == "False" {
				status.manage().MarkFalse(AppConditionServiceBindingsReady, "service binding %s failed: %v", binding.Name, cond.Reason)
				return
			}
			if cond.Status == "Unknown" {
				status.manage().MarkUnknown(AppConditionServiceBindingsReady, "service binding %s is not ready", binding.Name)
				return
			}
		}

	}

	status.manage().MarkTrue(AppConditionServiceBindingsReady)
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
