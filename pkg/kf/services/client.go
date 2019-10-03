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

package services

import (
	"context"
	"fmt"
	"time"

	cv1beta1 "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/typed/servicecatalog/v1beta1"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"knative.dev/pkg/apis"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	WaitForProvisionSuccess(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1beta1.ServiceInstance, err error)
}

// NewClient creates a new service client.
func NewClient(kclient cv1beta1.ServiceInstancesGetter) Client {
	return &coreClient{
		kclient: kclient,
	}
}

// WaitForProvisionSuccess is a utility function that combines WaitForE with ProvisionSuccess.
func (core *coreClient) WaitForProvisionSuccess(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1beta1.ServiceInstance, err error) {
	return core.WaitForE(ctx, namespace, name, interval, ProvisionSuccess)
}

// ProvisionSuccess implements ConditionFuncE and can be used to wait until an
// instance is successfully provisioned or fails.
func ProvisionSuccess(obj *v1beta1.ServiceInstance, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	// don't propagate old statuses
	if !ObservedGenerationMatchesGeneration(obj) {
		return false, nil
	}

	// Service catalog throws an additional check in, conditions are not in a
	// final state unless this is true.
	if obj.Status.AsyncOpInProgress {
		return false, nil
	}

	// Service catalog uses two independent conditions for success/failure rather
	// than a True/False on Ready. If Ready is false, it could be because the
	// instance is still provisioning. See the logic SC uses for their check for
	// more context:
	// https://github.com/kubernetes-sigs/service-catalog/blob/73229d43efccdff38e3a0265229c5814af24e0c3/pkg/svcat/service-catalog/instance.go#L266

	for _, cond := range ExtractConditions(obj) {
		if cond.Type == apis.ConditionType(v1beta1.ServiceBindingConditionFailed) && cond.IsTrue() {
			return true, fmt.Errorf("provision failed, message: %s reason: %s", cond.Message, cond.Reason)
		}

		if cond.Type == apis.ConditionType(v1beta1.ServiceBindingConditionReady) && cond.IsTrue() {
			return true, nil
		}
	}

	return false, nil
}
