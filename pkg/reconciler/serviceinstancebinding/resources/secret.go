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

package resources

import (
	"encoding/json"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// BindingCredentialsSecretName gets the name of the secret for a ServiceInstanceBinding.
func BindingCredentialsSecretName(binding *v1alpha1.ServiceInstanceBinding) string {
	return v1alpha1.GenerateName("credentials", binding.Name)
}

// MergeCredentialsSecretForBinding merges the Secret containing the binding credentials with params Secret.
func MergeCredentialsSecretForBinding(binding v1alpha1.ServiceInstanceBinding, serviceCredentialsSecret v1.Secret, bindingParamsSecret v1.Secret) (*v1.Secret, error) {
	var mergedCredentials map[string]json.RawMessage
	err := json.Unmarshal(serviceCredentialsSecret.Data[v1alpha1.ServiceInstanceParamsSecretKey], &mergedCredentials)
	if err != nil {
		return nil, err
	}

	// Merge in binding parameters, if there are any. Parameter values with the same key will override the original credential values.
	err = json.Unmarshal(bindingParamsSecret.Data[v1alpha1.ServiceInstanceParamsSecretKey], &mergedCredentials)
	if err != nil {
		return nil, err
	}

	newCreds := make(map[string][]byte)
	for k, v := range mergedCredentials {
		newCreds[k] = v
	}

	return makeBindingCredentialsSecret(&binding, newCreds), nil
}

// MakeCredentialsForOSBService creates a Secret containing the binding
// credentials for an OSB service.
func MakeCredentialsForOSBService(
	binding *v1alpha1.ServiceInstanceBinding,
	creds map[string]interface{},
) (*v1.Secret, error) {

	newCreds := make(map[string][]byte)
	for k, v := range creds {
		// Flatten OSB credentials into a Secret where keys are from OSB and
		// values are JSON encoded values so the creds can be properly consumed
		// by the code that builds VCAP_SERVICES and/or mounted directly into the
		// container.
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		newCreds[k] = encoded
	}

	return makeBindingCredentialsSecret(binding, newCreds), nil
}

func makeBindingCredentialsSecret(binding *v1alpha1.ServiceInstanceBinding, secretData map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BindingCredentialsSecretName(binding),
			Namespace: binding.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(binding),
			},
		},
		Data: secretData,
	}
}
