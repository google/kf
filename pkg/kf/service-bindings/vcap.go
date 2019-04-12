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

package servicebindings

import (
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// VcapServicesMap mimics CF's VCAP_SERVICES environment variable.
type VcapServicesMap map[string]VcapService

// Add inserts the VcapService to the map.
func (vm VcapServicesMap) Add(service VcapService) {
	vm[service.BindingName] = service
}

// VcapService represents a single entry in a VCAP_SERVICES map.
// It holds the credentials for a single service binding.
type VcapService struct {
	BindingName  string            `json:"binding_name"`
	InstanceName string            `json:"instance_name"`
	Name         string            `json:"name"`
	Plan         string            `json:"plan"`
	Credentials  map[string]string `json:"credentials"`
}

// NewVcapService creates a new VcapService given a binding and associated
// secret.
func NewVcapService(binding apiv1beta1.ServiceBinding, secret *corev1.Secret) VcapService {
	vs := VcapService{
		BindingName:  binding.Labels[BindingNameLabel],
		Name:         binding.Name,
		InstanceName: binding.Spec.InstanceRef.Name,
		Credentials:  make(map[string]string),
	}

	// Credentials are stored by the service catalog in a flat map, the data
	// values are just strings.
	for sn, sd := range secret.Data {
		vs.Credentials[sn] = string(sd)
	}

	// XXX: Plan is not populated but _is_ included in the struct for completeness.

	return vs
}
