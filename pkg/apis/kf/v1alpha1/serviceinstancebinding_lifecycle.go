// Copyright 2020 Google LLC
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
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *ServiceInstanceBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceInstanceBinding")
}

// PropagateParamsSecretStatus updates the status of the parameters secret being created and populated.
func (status *ServiceInstanceBindingStatus) PropagateParamsSecretStatus(secret *v1.Secret) {
	if secret == nil {
		status.manage().MarkUnknown(ServiceInstanceBindingConditionParamsSecretReady, "SecretMissing", "Secret for binding parameters doesn't exist")
		return
	}

	status.ParamsSecretCondition().MarkSuccess()

	// check if secret has been populated by the client
	contents, exists := secret.Data[ServiceInstanceBindingParamsSecretKey]
	switch {
	case !exists:
		// The Secret is created after the ServiceInstanceBinding so it can be owned
		// by the binding and will be deleted to prevent creds floating around.
		status.ParamsSecretPopulatedCondition().MarkUnknown(
			"SecretNotPopulated",
			"secret is missing key: %q",
			ServiceInstanceParamsSecretKey,
		)
	case !json.Valid(contents):
		status.ParamsSecretPopulatedCondition().MarkFalse(
			"SecretInvalid",
			"secret key %q must be valid JSON",
			ServiceInstanceParamsSecretKey,
		)
	default:
		status.ParamsSecretPopulatedCondition().MarkSuccess()
	}
}

// PropagateServiceFieldsStatus copies the service fields from the Service Instance status into the Service Binding status.
func (status *ServiceInstanceBindingStatus) PropagateVolumeStatus(serviceInstance *ServiceInstance, secret *v1.Secret) {
	if serviceInstance.Status.VolumeStatus == nil {
		// Not a Volume instance, mark as success.
		status.VolumeParamsPopulatedCondition().MarkSuccess()
		return
	}

	paramsJSON, ok := secret.Data[ServiceInstanceBindingParamsSecretKey]
	if !ok {
		status.VolumeParamsPopulatedCondition().MarkFalse("Secret is missing key %q", ServiceInstanceBindingParamsSecretKey)
		return
	}

	params := BindingVolumeParams{}
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		status.VolumeParamsPopulatedCondition().MarkFalse("failed to unmarshal params from Secret: %v", err.Error())
		return
	}

	// Validate UID and GID
	if params.UID != "" {
		if _, err := params.UIDInt64(); err != nil {
			status.VolumeParamsPopulatedCondition().MarkFalse("Invalid UID", "UID needs to be non-negative integers")
			return
		}
	}

	if params.GID != "" {
		if _, err := params.GIDInt64(); err != nil {
			status.VolumeParamsPopulatedCondition().MarkFalse("Invalid GID ", "GID need to be non-negative integers")
			return
		}
	}

	if params.Mount == "" {
		status.VolumeParamsPopulatedCondition().MarkFalse("mount missing", "Mount Path is required for VolumeBindings")
		return
	}

	status.VolumeStatus = &BindingVolumeStatus{
		Mount:                     params.Mount,
		PersistentVolumeName:      serviceInstance.Status.VolumeStatus.PersistentVolumeName,
		PersistentVolumeClaimName: serviceInstance.Status.VolumeStatus.PersistentVolumeClaimName,
		ReadOnly:                  params.ReadOnly,
		UidGid: UidGid{
			UID: params.UID,
			GID: params.GID,
		},
	}

	status.VolumeParamsPopulatedCondition().MarkSuccess()
}

// PropagateCredentialsSecretStatus updates the status of the secret holding the credentials for the service instance binding.
func (status *ServiceInstanceBindingStatus) PropagateCredentialsSecretStatus(secret *v1.Secret) {
	if secret == nil {
		status.CredentialsSecretRef.Name = ""
		status.manage().MarkUnknown(ServiceInstanceBindingConditionCredentialsSecretReady, "SecretMissing", "Secret for binding credentials doesn't exist")
		return
	}
	status.CredentialsSecretRef.Name = secret.Name
	status.CredentialsSecretCondition().MarkSuccess()
}

// PropagateServiceFieldsStatus copies the service fields from the Service Instance status into the Service Binding status.
func (status *ServiceInstanceBindingStatus) PropagateServiceFieldsStatus(serviceInstance *ServiceInstance) {
	status.ServiceFields = *serviceInstance.Status.ServiceFields.DeepCopy()
}

// PropagateServiceInstanceStatus propagates the Service Instance status to the Service Binding status.
func (status *ServiceInstanceBindingStatus) PropagateServiceInstanceStatus(serviceInstance *ServiceInstance) {
	cond := serviceInstance.Status.GetCondition(ServiceInstanceConditionReady)
	PropagateCondition(status.manage(), ServiceInstanceBindingConditionServiceInstanceReady, cond)
}

// PropagateBindingNameStatus propagates the binding name to the Service Instance Binding status.
func (status *ServiceInstanceBindingStatus) PropagateBindingNameStatus(binding *ServiceInstanceBinding) {
	bindingOverride := binding.Spec.BindingNameOverride
	if bindingOverride != "" {
		status.BindingName = bindingOverride
	} else {
		status.BindingName = binding.Spec.InstanceRef.Name
	}
}

// PropagateRouteServiceURLStatus copies the route service URL from the Service Instance status into the Service Binding status.
func (status *ServiceInstanceBindingStatus) PropagateRouteServiceURLStatus(serviceInstance *ServiceInstance) {
	status.RouteServiceURL = serviceInstance.Status.RouteServiceURL.DeepCopy()
}

// MarkBackingResourceReady notes that the backing resource is ready.
// This is always true if no backing resource exists.
func (status *ServiceInstanceBindingStatus) MarkBackingResourceReady() {
	status.BackingResourceCondition().MarkSuccess()
}

// PropagateUnbindStatus propagates the result of an OSB unbind request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceBindingStatus) PropagateUnbindStatus(
	response *osbclient.UnbindResponse,
	err error,
) {
	// This isn't public because we shouldn't rely on any particular value.
	const reasonUnbinding = "Unbinding"

	condition := status.BackingResourceCondition()

	switch {
	case isDeletedOSBError(err):
		// If the resource is already gone, mark as deleted.
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Unbound: &OSBState{},
		}

	case err != nil:
		condition.MarkReconciliationError(reasonUnbinding, err)
		status.OSBStatus = BindingOSBStatus{
			UnbindFailed: &OSBState{},
		}

	case response.Async:
		condition.MarkUnknown(reasonUnbinding, "operation is pending")
		status.OSBStatus = BindingOSBStatus{
			Unbinding: &OSBState{
				OperationKey: (*string)(response.OperationKey),
			},
		}

	default:
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Unbound: &OSBState{},
		}
	}
}

// PropagateUnbindLastOperationStatus propagates the result of an asynchronous
// OSB unbind request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceBindingStatus) PropagateUnbindLastOperationStatus(
	response *osbclient.LastOperationResponse,
	err error,
) {

	condition := status.BackingResourceCondition()

	switch {
	case isRetryableOSBError(err):
		condition.MarkUnknown(
			"PollingOperation",
			"temporary error while polling: %v",
			err,
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Unbinding status for the state to get here.

	case isDeletedOSBError(err):
		// If the resource is already gone, mark as deleted.
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Unbound: &OSBState{},
		}

	case err != nil:
		condition.MarkReconciliationError("PollingOperation", err)
		status.OSBStatus = BindingOSBStatus{
			UnbindFailed: &OSBState{},
		}

	case osbclient.StateInProgress == response.State:
		condition.MarkUnknown(
			"UnbindingAsync",
			formatOperationMessage(response),
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Unbinding status for the state to get here.

	case osbclient.StateSucceeded == response.State:
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Unbound: &OSBState{},
		}

	case osbclient.StateFailed == response.State:
		condition.MarkReconciliationError(
			"UnbindFailed",
			fmt.Errorf("unbind failed: %s", formatOperationMessage(response)),
		)
		status.OSBStatus = BindingOSBStatus{
			UnbindFailed: &OSBState{},
		}

	default:
		condition.MarkReconciliationError("UnknownState",
			fmt.Errorf("unknown state: %s", formatOperationMessage(response)))
		status.OSBStatus = BindingOSBStatus{
			UnbindFailed: &OSBState{},
		}
	}
}

// PropagateBindStatus propagates the result of an OSB bind request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceBindingStatus) PropagateBindStatus(
	response *osbclient.BindResponse,
	err error,
) {
	condition := status.BackingResourceCondition()

	switch {
	case err != nil:
		condition.MarkReconciliationError("Binding", err)
		status.OSBStatus = BindingOSBStatus{
			BindFailed: &OSBState{},
		}

	case response.Async:
		condition.MarkUnknown("BindingAsync", "operation is pending")
		status.OSBStatus = BindingOSBStatus{
			Binding: &OSBState{
				OperationKey: (*string)(response.OperationKey),
			},
		}

	default:
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Bound: &OSBState{},
		}
	}
}

// PropagateBindLastOperationStatus propagates the result of an asynchronous
// OSB bind request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceBindingStatus) PropagateBindLastOperationStatus(
	response *osbclient.LastOperationResponse,
	err error,
) {

	// This isn't public because we shouldn't rely on any particular value.
	const reasonBindingAsync = "BindingAsync"

	condition := status.BackingResourceCondition()
	switch {
	case isRetryableOSBError(err):
		condition.MarkUnknown(
			reasonBindingAsync,
			"temporary error while polling: %v",
			err,
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Binding status for the state to get here.

	case err != nil:
		condition.MarkReconciliationError("PollingOperation", err)
		status.OSBStatus = BindingOSBStatus{
			BindFailed: &OSBState{},
		}

	case osbclient.StateInProgress == response.State:
		condition.MarkUnknown(
			reasonBindingAsync,
			formatOperationMessage(response),
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Binding status for the state to get here.

	case osbclient.StateSucceeded == response.State:
		condition.MarkSuccess()
		status.OSBStatus = BindingOSBStatus{
			Bound: &OSBState{},
		}

	case osbclient.StateFailed == response.State:
		condition.MarkReconciliationError(
			"BindFailed",
			fmt.Errorf("bind failed: %s", formatOperationMessage(response)),
		)
		status.OSBStatus = BindingOSBStatus{
			BindFailed: &OSBState{},
		}

	default:
		condition.MarkReconciliationError("UnknownState",
			fmt.Errorf("unknown state: %s", formatOperationMessage(response)))
		status.OSBStatus = BindingOSBStatus{
			BindFailed: &OSBState{},
		}
	}
}
