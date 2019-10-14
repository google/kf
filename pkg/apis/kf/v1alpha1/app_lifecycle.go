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

	serving "github.com/google/kf/third_party/knative-serving/pkg/apis/serving/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
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

// IsReady looks at the conditions to see if they are happy.
func (status *AppStatus) IsReady() bool {
	return status.manage().IsHappy()
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
	// Stopped apps don't have a Knative service
	if service == nil {
		return
	}

	if service.Status.ObservedGeneration != service.Generation {
		status.manage().MarkUnknown(AppConditionKnativeServiceReady, "GenerationMismatch", "the Knative service needs to be synchronized")
		return
	}

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

// PropagateRouteStatus updates the route readiness status.
func (status *AppStatus) PropagateRouteStatus() {
	status.manage().MarkTrue(AppConditionRouteReady)
}

// serviceBindingConditionType creates a Conditiontype for a ServiceBinding.
func serviceBindingConditionType(binding *servicecatalogv1beta1.ServiceBinding) (apis.ConditionType, error) {
	serviceInstance, ok := binding.Labels[ComponentLabel]
	if !ok {
		return "", fmt.Errorf("binding %s is missing the label %s", binding.Name, ComponentLabel)
	}

	return apis.ConditionType(fmt.Sprintf("%sReady", serviceInstance)), nil
}

// PropagateServiceBindingsStatus updates the service binding readiness status.
func (status *AppStatus) PropagateServiceBindingsStatus(bindings []servicecatalogv1beta1.ServiceBinding) {

	// Gather binding names
	var bindingNames []string
	for _, binding := range bindings {
		bindingNames = append(bindingNames, binding.Name)
	}
	status.ServiceBindingNames = bindingNames

	// Gather binding conditions
	var conditionTypes []apis.ConditionType
	for _, binding := range bindings {
		conditionType, err := serviceBindingConditionType(&binding)
		if err != nil {
			status.manage().MarkFalse(AppConditionServiceBindingsReady, "Error", "%v", err)
			return
		}

		conditionTypes = append(conditionTypes, conditionType)
	}

	duckStatus := &duckv1beta1.Status{}
	manager := apis.NewLivingConditionSet(conditionTypes...).Manage(duckStatus)
	manager.InitializeConditions()

	for _, binding := range bindings {
		if binding.Generation != binding.Status.ReconciledGeneration {

			// this binding's conditions are out of date.
			continue
		}

		for _, cond := range binding.Status.Conditions {
			if cond.Type != servicecatalogv1beta1.ServiceBindingConditionReady {
				continue
			}

			conditionType, err := serviceBindingConditionType(&binding)
			if err != nil {
				status.manage().MarkFalse(AppConditionServiceBindingsReady, "Error", "%v", err)
				return
			}
			switch v1.ConditionStatus(cond.Status) {
			case v1.ConditionTrue:
				manager.MarkTrue(conditionType)
			case v1.ConditionFalse:
				manager.MarkFalse(conditionType, cond.Reason, cond.Message)
			case v1.ConditionUnknown:
				manager.MarkUnknown(conditionType, cond.Reason, cond.Message)
			}
		}
	}

	// if there are no bindings, set the happy condition to true
	if len(bindings) == 0 {
		manager.MarkTrue(apis.ConditionReady)
	}

	// Copy Ready condition
	PropagateCondition(status.manage(), AppConditionServiceBindingsReady, manager.GetCondition(apis.ConditionReady))
	status.ServiceBindingConditions = duckStatus.Conditions
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

func (status *AppStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
