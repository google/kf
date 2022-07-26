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

package resources

import (
	"errors"
	"fmt"

	"encoding/json"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/ptr"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

// MakeOSBDeprovisionRequest creates a DeprovisionRequest for OSB backed services.
func MakeOSBDeprovisionRequest(serviceInstance *v1alpha1.ServiceInstance) *osbclient.DeprovisionRequest {
	return &osbclient.DeprovisionRequest{
		InstanceID:        fmt.Sprintf("%s", serviceInstance.UID),
		AcceptsIncomplete: true,
		ServiceID:         serviceInstance.Spec.OSB.ClassUID,
		PlanID:            serviceInstance.Spec.OSB.PlanUID,

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}
}

// MakeOSBLastOperationRequest creates a LastOperationRequest for OSB backed services,
// used to poll long running operations.
func MakeOSBLastOperationRequest(serviceInstance *v1alpha1.ServiceInstance, operationKey *string) *osbclient.LastOperationRequest {
	return &osbclient.LastOperationRequest{
		InstanceID:   fmt.Sprintf("%s", serviceInstance.UID),
		ServiceID:    ptr.String(serviceInstance.Spec.OSB.ClassUID),
		PlanID:       ptr.String(serviceInstance.Spec.OSB.PlanUID),
		OperationKey: (*osbclient.OperationKey)(operationKey),

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}
}

// MakeOSBProvisionRequest creates a request to provision an OSB resource.
func MakeOSBProvisionRequest(
	serviceInstance *v1alpha1.ServiceInstance,
	namespace *corev1.Namespace,
	paramsSecret *corev1.Secret,
) (*osbclient.ProvisionRequest, error) {
	if serviceInstance == nil || namespace == nil || paramsSecret == nil {
		return nil, errors.New("ServiceInstance, Namespace, and Secret are all required")
	}

	paramsJSON, ok := paramsSecret.Data[v1alpha1.ServiceInstanceParamsSecretKey]
	if !ok {
		return nil, fmt.Errorf("Secret was missing key %q", v1alpha1.ServiceInstanceParamsSecretKey)
	}

	params := make(map[string]interface{})
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal params from Secret: %s", err.Error())
	}

	namespaceUID := fmt.Sprintf("%s", namespace.UID)
	return &osbclient.ProvisionRequest{
		// Use the UID for tracibility and to ensure duplicate requests (if any)
		// only get provisioned once.
		InstanceID:        fmt.Sprintf("%s", serviceInstance.UID),
		AcceptsIncomplete: true,
		ServiceID:         serviceInstance.Spec.OSB.ClassUID,
		PlanID:            serviceInstance.Spec.OSB.PlanUID,
		// We're treating the organization and space as the same entity.
		OrganizationGUID: namespaceUID,
		SpaceGUID:        namespaceUID,
		Parameters:       params,
		Context:          CreateOSBContext(serviceInstance, namespace),

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}, nil
}

// CreateOSBContext creates a context object for OSB requests.
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#context-object
func CreateOSBContext(
	serviceInstance *v1alpha1.ServiceInstance,
	namespace *corev1.Namespace,
) map[string]interface{} {

	namespaceUID := fmt.Sprintf("%s", namespace.UID)

	return map[string]interface{}{
		"platform":      "kf",
		"instance_name": serviceInstance.Name,

		// CF Style context properties from:
		// https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#cloud-foundry-context-object
		// We're treating the organization and space as the same entity.
		"organization_guid": namespaceUID,
		"organization_name": namespace.Name,
		"space_guid":        namespaceUID,
		"space_name":        namespace.Name,

		// Don't include annotations, it's a recent addition
		// to OSB for CF, but in Kubernetes, that would be a
		// security concern due to annotations being used
		// differently e.g. holding entire dumps of kubectl
		// apply.

		// Kubernetes style context properties from:
		// https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#kubernetes-context-object
		// these values help Minibroker.
		"namespace": serviceInstance.Namespace,
	}
}
