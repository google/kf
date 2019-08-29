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

package cfutil

import (
	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// VcapServicesMap mimics CF's VCAP_SERVICES environment variable.
// See https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#VCAP-SERVICES
// for more information about the structure.
// The key for each service is the same as the value of the "label" attribute.
type VcapServicesMap map[string][]VcapService

// Add inserts the VcapService to the map.
func (vm VcapServicesMap) Add(service VcapService) {
	// See the cloud-controller-ng source for the definition of how this is
	// built.
	// https://github.com/cloudfoundry/cloud_controller_ng/blob/65a75e6c97f49756df96e437e253f033415b2db1/app/presenters/system_environment/system_env_presenter.rb
	vm[service.Label] = append(vm[service.Label], service)
}

// VcapService represents a single entry in a VCAP_SERVICES map.
// It holds the credentials for a single service binding.
type VcapService struct {
	BindingName  string            `json:"binding_name"`  // The name assigned to the service binding by the user.
	InstanceName string            `json:"instance_name"` // The name assigned to the service instance by the user.
	Name         string            `json:"name"`          // The binding_name if it exists; otherwise the instance_name.
	Label        string            `json:"label"`         // The name of the service offering.
	Tags         []string          `json:"tags"`          // An array of strings an app can use to identify a service instance.
	Plan         string            `json:"plan"`          // The service plan selected when the service instance was created.
	Credentials  map[string]string `json:"credentials"`   // The service-specific credentials needed to access the service instance.
}

// NewVcapService creates a new VcapService given a binding and associated
// secret.
func NewVcapService(instance apiv1beta1.ServiceInstance, binding apiv1beta1.ServiceBinding, secret *corev1.Secret) VcapService {
	// See the cloud-controller-ng source for how this is supposed to be built
	// being that it doesn't seem to be formally fully documented anywhere:
	// https://github.com/cloudfoundry/cloud_controller_ng/blob/65a75e6c97f49756df96e437e253f033415b2db1/app/presenters/system_environment/service_binding_presenter.rb#L32
	vs := VcapService{
		BindingName:  binding.Labels[kfv1alpha1.ComponentLabel],
		Name:         coalesce(binding.Labels[kfv1alpha1.ComponentLabel], binding.Spec.InstanceRef.Name),
		InstanceName: binding.Spec.InstanceRef.Name,
		Label:        coalesce(instance.Spec.ServiceClassExternalName, instance.Spec.ClusterServiceClassExternalName),
		Plan:         coalesce(instance.Spec.ServicePlanExternalName, instance.Spec.ClusterServicePlanExternalName),
		Credentials:  make(map[string]string),
	}

	// TODO(josephlewis42) we need to get tags from the (Cluster)ServiceClass
	// this could be aided by the BindingParentHierarchy function in the
	// service catalog SDK.

	// Credentials are stored by the service catalog in a flat map, the data
	// values are just strings.
	for sn, sd := range secret.Data {
		vs.Credentials[sn] = string(sd)
	}

	return vs
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}

	return ""
}
