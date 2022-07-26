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
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *ServiceInstance) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceInstance")
}

// PropagateSecretStatus updates the status of the parameters secret being created and populated.
func (status *ServiceInstanceStatus) PropagateSecretStatus(secret *corev1.Secret) {
	if secret == nil {
		status.manage().MarkUnknown(ServiceInstanceConditionParamsSecretReady, "SecretMissing", "Secret for instance parameters does not exist")
		return
	}

	status.SecretName = secret.Name
	status.ParamsSecretCondition().MarkSuccess()

	// check if secret has been populated by the client
	contents, exists := secret.Data[ServiceInstanceParamsSecretKey]
	switch {
	case !exists:
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

// PropagateServiceFieldsStatus copies the service fields from the Service Instance Spec into the Status.
func (status *ServiceInstanceStatus) PropagateServiceFieldsStatus(serviceInstance *ServiceInstance) {
	status.ServiceFields.Tags = serviceInstance.Spec.Tags
	switch {
	case serviceInstance.IsUserProvided():
		status.ServiceTypeDescription = UserProvidedServiceDescription
		status.ServiceFields.ClassName = serviceInstance.Spec.UPS.MockClassName
		status.ServiceFields.PlanName = serviceInstance.Spec.UPS.MockPlanName

		if status.ServiceFields.ClassName == "" {
			status.ServiceFields.ClassName = UserProvidedServiceClassName
		}
	case serviceInstance.IsLegacyBrokered():
		status.ServiceTypeDescription = BrokeredServiceDescription
		status.ServiceFields.ClassName = serviceInstance.Spec.Brokered.ClassName
		status.ServiceFields.PlanName = serviceInstance.Spec.Brokered.PlanName
	case serviceInstance.IsKfBrokered():
		status.ServiceTypeDescription = BrokeredServiceDescription
		status.ServiceFields.ClassName = serviceInstance.Spec.OSB.ClassName
		status.ServiceFields.PlanName = serviceInstance.Spec.OSB.PlanName
	case serviceInstance.IsVolume():
		status.ServiceTypeDescription = VolumeServiceDescription
		status.ServiceFields.ClassName = serviceInstance.Spec.Volume.ClassName
		status.ServiceFields.PlanName = serviceInstance.Spec.Volume.PlanName
	}
}

// PropagateRouteServiceURLStatus copies the route service URL from the Service Instance Spec into the Status.
func (status *ServiceInstanceStatus) PropagateRouteServiceURLStatus(serviceInstance *ServiceInstance) {
	if serviceInstance.IsRouteService() {
		status.RouteServiceURL = serviceInstance.Spec.UPS.RouteServiceURL.DeepCopy()
	} else {
		status.RouteServiceURL = nil
	}
}

// PropagateVolumeServiceStatus copies the k8s volume objects names into the Status.
func (status *ServiceInstanceStatus) PropagateVolumeServiceStatus(serviceInstance *ServiceInstance, volumeName, volumeClaimName string) {
	if serviceInstance.IsVolume() {
		status.VolumeStatus = &VolumeStatus{
			PersistentVolumeName:      volumeName,
			PersistentVolumeClaimName: volumeClaimName,
		}
	} else {
		status.VolumeStatus = nil
	}
}

// PropagateDeploymentStatus updates the deployment status to reflect the
// underlying state of the deployment.
func (status *ServiceInstanceStatus) PropagateDeploymentStatus(deployment *appsv1.Deployment) {
	for _, cond := range deployment.Status.Conditions {
		// ReplicaFailure is added in a deployment when one of its pods fails to be created
		// or deleted.
		if cond.Type == appsv1.DeploymentReplicaFailure && cond.Status == corev1.ConditionTrue {
			status.manage().MarkFalse(ServiceInstanceConditionBackingResourceReady, cond.Reason, cond.Message)
			return
		}
	}

	if deployment.Generation > deployment.Status.ObservedGeneration {
		status.manage().MarkUnknown(ServiceInstanceConditionBackingResourceReady, "GenerationOutOfDate", fmt.Sprintf("waiting for deployment spec update to be observed"))
		return
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing && cond.Reason == "ProgressDeadlineExceeded" {
			status.manage().MarkFalse(ServiceInstanceConditionBackingResourceReady, "DeadlineExceeded", fmt.Sprintf("deployment %q exceeded its progress deadline", deployment.Name))
			return
		}
	}

	if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
		status.manage().MarkUnknown(ServiceInstanceConditionBackingResourceReady, "UpdatingReplicas", fmt.Sprintf("waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas))
		return
	}
	if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		status.manage().MarkUnknown(ServiceInstanceConditionBackingResourceReady, "TerminatingOldReplicas", fmt.Sprintf("waiting for deployment %q rollout to finish: %d old replicas are pending termination", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas))
		return
	}
	if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		status.manage().MarkUnknown(ServiceInstanceConditionBackingResourceReady, "InitializingPods", fmt.Sprintf("waiting for deployment %q rollout to finish: %d of %d updated replicas are available", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas))
		return
	}

	status.BackingResourceCondition().MarkSuccess()
}

// PropagateDeletionBlockedStatus updates the ready status of the service instance to False if the service received a delete request
// and is still part of a service binding.
func (status *ServiceInstanceStatus) PropagateDeletionBlockedStatus() {
	status.manage().MarkFalse(ServiceInstanceConditionReady, "DeletionBlocked", "ServiceInstance is part of a service binding")
}

// MarkSpaceHealthy notes that the space was able to be retrieved and
// defaults can be applied from it.
func (status *ServiceInstanceStatus) MarkSpaceHealthy() {
	status.manage().MarkTrue(ServiceInstanceConditionSpaceReady)
}

// MarkSpaceUnhealthy notes that the space was could not be retrieved.
func (status *ServiceInstanceStatus) MarkSpaceUnhealthy(reason, message string) {
	status.manage().MarkFalse(ServiceInstanceConditionSpaceReady, reason, message)
}

// MarkBackingResourceReady notes that the backing resource is ready.
// This is always true if no backing resource exists.
func (status *ServiceInstanceStatus) MarkBackingResourceReady() {
	status.manage().MarkTrue(ServiceInstanceConditionBackingResourceReady)
}

// PropagateDeprovisionStatus propagates the result of a synchronous
// OSB deprovision request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceStatus) PropagateDeprovisionStatus(
	response *osbclient.DeprovisionResponse,
	err error,
) {
	condition := status.BackingResourceCondition()

	switch {
	case isDeletedOSBError(err):
		// If the resource is already gone, mark as deleted.
		condition.MarkSuccess()
		status.OSBStatus = OSBStatus{
			Deprovisioned: &OSBState{},
		}

	case err != nil:
		status.OSBStatus = OSBStatus{
			DeprovisionFailed: &OSBState{},
		}
		condition.MarkReconciliationError("DeprovisioningInstance", err)

	case response.Async:
		status.OSBStatus = OSBStatus{
			Deprovisioning: &OSBState{
				OperationKey: (*string)(response.OperationKey),
			},
		}
		condition.MarkUnknown("DeprovisioningInstance", "operation is pending")

	default:
		status.OSBStatus = OSBStatus{
			Deprovisioned: &OSBState{},
		}
		condition.MarkSuccess()
	}
}

// PropagateDeprovisionAsyncStatus propagates the result of an asynchronous
// OSB deprovision request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceStatus) PropagateDeprovisionAsyncStatus(
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
		// contain the Deprovisioning status for the state to get here.

	case isDeletedOSBError(err):
		// If the resource is already gone, mark as deleted.
		condition.MarkSuccess()
		status.OSBStatus = OSBStatus{
			Deprovisioned: &OSBState{},
		}

	case err != nil:
		condition.MarkReconciliationError("PollingOperation", err)
		status.OSBStatus = OSBStatus{
			DeprovisionFailed: &OSBState{},
		}

	case osbclient.StateInProgress == response.State:
		condition.MarkUnknown(
			"DeprovisioningInstance",
			formatOperationMessage(response),
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Deprovisioning status for the state to get here.

	case osbclient.StateSucceeded == response.State:
		condition.MarkSuccess()
		status.OSBStatus = OSBStatus{
			Deprovisioned: &OSBState{},
		}

	case osbclient.StateFailed == response.State:
		condition.MarkReconciliationError(
			"DeprovisionFailed",
			fmt.Errorf("deprovision failed: %s", formatOperationMessage(response)),
		)
		status.OSBStatus = OSBStatus{
			DeprovisionFailed: &OSBState{},
		}

	default:
		condition.MarkReconciliationError("UnknownState",
			fmt.Errorf("unknown state: %s", formatOperationMessage(response)))
		status.OSBStatus = OSBStatus{
			DeprovisionFailed: &OSBState{},
		}
	}
}

// PropagateProvisionStatus propagates the result of a synchronous
// OSB provision request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceStatus) PropagateProvisionStatus(
	response *osbclient.ProvisionResponse,
	err error,
) {
	condition := status.BackingResourceCondition()

	switch {
	case err != nil:
		condition.MarkReconciliationError(
			"ProvisioningInstance",
			fmt.Errorf("couldn't provision: %v", err),
		)
		status.OSBStatus = OSBStatus{
			ProvisionFailed: &OSBState{},
		}

	case response.Async:
		status.OSBStatus = OSBStatus{
			Provisioning: &OSBState{
				OperationKey: (*string)(response.OperationKey),
			},
		}
		condition.MarkUnknown("ProvisioningInstance", "operation is pending")

	default:
		status.OSBStatus = OSBStatus{
			Provisioned: &OSBState{},
		}
		condition.MarkSuccess()
	}
}

// PropagateProvisionAsyncStatus propagates the result of an asynchronous
// OSB provision request.
//
// At the end of this call, the backing resource condition and OSBStatus field
// will be updated.
func (status *ServiceInstanceStatus) PropagateProvisionAsyncStatus(
	response *osbclient.LastOperationResponse,
	err error,
) {
	condition := status.BackingResourceCondition()
	switch {
	case isRetryableOSBError(err):
		condition.MarkUnknown(
			"ProvisioningInstance",
			"temporary error while polling: %v",
			err,
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Provisioning status for the state to get here.

	case err != nil:
		condition.MarkReconciliationError("PollingOperation", err)
		status.OSBStatus = OSBStatus{
			ProvisionFailed: &OSBState{},
		}

	case osbclient.StateInProgress == response.State:
		condition.MarkUnknown(
			"ProvisioningInstance",
			formatOperationMessage(response),
		)
		// No update is necessary to OSBStatus, it should already
		// contain the Provisioning status for the state to get here.

	case osbclient.StateSucceeded == response.State:
		condition.MarkSuccess()
		status.OSBStatus = OSBStatus{
			Provisioned: &OSBState{},
		}

	case osbclient.StateFailed == response.State:
		condition.MarkReconciliationError(
			"ProvisionFailed",
			fmt.Errorf("provision failed: %s", formatOperationMessage(response)),
		)
		status.OSBStatus = OSBStatus{
			ProvisionFailed: &OSBState{},
		}

	default:
		condition.MarkReconciliationError("UnknownState",
			fmt.Errorf("unknown state: %s", formatOperationMessage(response)))
		status.OSBStatus = OSBStatus{
			ProvisionFailed: &OSBState{},
		}
	}
}

func isRetryableOSBError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := osbclient.IsHTTPError(err); ok {
		return httpErr.StatusCode == http.StatusConflict ||
			httpErr.StatusCode >= 500
	}

	return false
}

func isDeletedOSBError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := osbclient.IsHTTPError(err); ok {
		return httpErr.StatusCode == http.StatusGone ||
			httpErr.StatusCode == http.StatusNotFound
	}

	return false
}

func formatOperationMessage(op *osbclient.LastOperationResponse) string {
	if op == nil {
		return "(nil operation)"
	}

	message := fmt.Sprintf("operation state: %q", op.State)
	if op.Description != nil {
		message = fmt.Sprintf("%s description: %q", message, *op.Description)
	}

	return message
}
