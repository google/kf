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

package v1alpha1

import corev1 "k8s.io/api/core/v1"

// PropagateDeletionBlockedStatus updates the ready status of the service broker
// to False if the broker received a delete request and is still part of a
// service instance.
func (status *CommonServiceBrokerStatus) PropagateDeletionBlockedStatus() {
	status.manage().MarkFalse(
		CommonServiceBrokerConditionReady,
		"DeletionBlocked",
		"broker is being used by one or more service instances",
	)
}

// PropagateSecretStatus updates the status of the parameters secret being created and populated.
func (status *CommonServiceBrokerStatus) PropagateSecretStatus(secret *corev1.Secret) {
	if secret == nil {
		status.CredsSecretCondition().
			MarkUnknown("SecretMissing", "Broker credentials Secret doesn't exist")
		return
	}

	status.CredsSecretCondition().MarkSuccess()

	// check if secret has been populated by the client, but not the validity
	if len(secret.Data) > 0 {
		status.CredsSecretPopulatedCondition().MarkSuccess()
	} else {
		status.CredsSecretPopulatedCondition().
			MarkUnknown("SecretNotPopulated", "Secret value for parameters does not exist")
	}
}
