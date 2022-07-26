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
	serviceinstanceresources "github.com/google/kf/v2/pkg/reconciler/serviceinstance/resources"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/ptr"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

// MakeOSBUnbindRequest creates an UnbindRequest for OSB backed services.
func MakeOSBUnbindRequest(
	serviceInstance *v1alpha1.ServiceInstance,
	binding *v1alpha1.ServiceInstanceBinding,
) *osbclient.UnbindRequest {

	return &osbclient.UnbindRequest{
		InstanceID:        fmt.Sprintf("%s", serviceInstance.UID),
		BindingID:         fmt.Sprintf("%s", binding.UID),
		AcceptsIncomplete: true,
		ServiceID:         serviceInstance.Spec.OSB.ClassUID,
		PlanID:            serviceInstance.Spec.OSB.PlanUID,

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}
}

// MakeOSBBindingLastOperationRequest creates a LastOperationRequest for OSB backed services,
// used to poll long running operations.
func MakeOSBBindingLastOperationRequest(
	serviceInstance *v1alpha1.ServiceInstance,
	binding *v1alpha1.ServiceInstanceBinding,
	operationKey *string,
) *osbclient.BindingLastOperationRequest {

	return &osbclient.BindingLastOperationRequest{
		InstanceID:   fmt.Sprintf("%s", serviceInstance.UID),
		BindingID:    fmt.Sprintf("%s", binding.UID),
		ServiceID:    ptr.String(serviceInstance.Spec.OSB.ClassUID),
		PlanID:       ptr.String(serviceInstance.Spec.OSB.PlanUID),
		OperationKey: (*osbclient.OperationKey)(operationKey),

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}
}

// MakeOSBBindRequest creates a request to bind to an OSB resource.
func MakeOSBBindRequest(
	serviceInstance *v1alpha1.ServiceInstance,
	binding *v1alpha1.ServiceInstanceBinding,
	namespace *corev1.Namespace,
	paramsSecret *corev1.Secret,
) (*osbclient.BindRequest, error) {

	if serviceInstance == nil || namespace == nil || paramsSecret == nil || binding == nil {
		return nil, errors.New("ServiceInstance, ServiceInstanceBinding, Namespace, and Secret are all required")
	}

	paramsJSON, ok := paramsSecret.Data[v1alpha1.ServiceInstanceBindingParamsSecretKey]
	if !ok {
		return nil, fmt.Errorf("Secret was missing key %q", v1alpha1.ServiceInstanceBindingParamsSecretKey)
	}

	params := make(map[string]interface{})
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal params from Secret: %s", err.Error())
	}

	return &osbclient.BindRequest{
		// Use the UID for tracibility and to ensure duplicate requests (if any)
		// only get provisioned once.
		InstanceID:        fmt.Sprintf("%s", serviceInstance.UID),
		BindingID:         fmt.Sprintf("%s", binding.UID),
		AcceptsIncomplete: true,
		ServiceID:         serviceInstance.Spec.OSB.ClassUID,
		PlanID:            serviceInstance.Spec.OSB.PlanUID,
		Parameters:        params,
		Context:           serviceinstanceresources.CreateOSBContext(serviceInstance, namespace),

		// Don't send OriginatingIdentity to the broker which may include
		// PII (user's GAIA ID, or Project ID).
	}, nil
}

// MakeOSBGetBindingRequest creates a GetBindingRequest for OSB backed services,
// used to poll long running operations.
func MakeOSBGetBindingRequest(
	serviceInstance *v1alpha1.ServiceInstance,
	binding *v1alpha1.ServiceInstanceBinding,
) *osbclient.GetBindingRequest {

	return &osbclient.GetBindingRequest{
		InstanceID: fmt.Sprintf("%s", serviceInstance.UID),
		BindingID:  fmt.Sprintf("%s", binding.UID),
	}
}
