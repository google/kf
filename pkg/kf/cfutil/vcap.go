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
	"encoding/json"

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
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
// For user-provided services, many fields are omitted if they are empty, such as binding name, instance name, and plan.
// If there are no tags for the service, the JSON encoding is an empty list rather than null.
// See http://engineering.pivotal.io/post/spring-boot-injecting-credentials/
// for an example of VCAP_SERVICES for a user-provided service.
type VcapService struct {
	BindingName  *string                    `json:"binding_name,omitempty"`  // The name assigned to the service binding by the user.
	InstanceName string                     `json:"instance_name,omitempty"` // The name assigned to the service instance by the user.
	Name         string                     `json:"name"`                    // The binding_name if it exists; otherwise the instance_name.
	Label        string                     `json:"label"`                   // The name of the service offering.
	Tags         []string                   `json:"tags"`                    // An array of strings an app can use to identify a service instance.
	Plan         string                     `json:"plan,omitempty"`          // The service plan selected when the service instance was created.
	Credentials  map[string]json.RawMessage `json:"credentials"`             // The service-specific credentials needed to access the service instance.
	VolumeMounts []VolumeMount              `json:"volume_mounts,omitempty"` // Only for VolumeServiceBinding. The volume specific information.
}

// VolumeMount contains volume specific information.
type VolumeMount struct {
	// ContainerDir contains the path to the mounted volume that you bound to your App.
	ContainerDir string `json:"container_dir,omitempty"`
	// DeviceType is the NFS volume release. This currently only supports shared devices.
	// A shared device represents a distributed file system that can mount on all App instances simultaneously.
	DeviceType string `json:"device_type,omitempty"`
	// Mode is a string that informs what type of access your App has to NFS, either read-only (`ro`) or read-write (`rw`).
	Mode string `json:"mode,omitempty"`
}

// NewVcapService creates a new VcapService given a binding and associated secret.
func NewVcapService(binding kfv1alpha1.ServiceInstanceBinding, credentialsSecret corev1.Secret) VcapService {
	// See the cloud-controller-ng source for how this is supposed to be built
	// being that it doesn't seem to be formally fully documented anywhere:
	// https://github.com/cloudfoundry/cloud_controller_ng/blob/65a75e6c97f49756df96e437e253f033415b2db1/app/presenters/system_environment/service_binding_presenter.rb#L32

	vs := VcapService{
		BindingName:  getBindingName(binding),
		Name:         binding.Status.BindingName,
		InstanceName: binding.Spec.InstanceRef.Name,
		Label:        binding.Status.ClassName,
		Plan:         binding.Status.PlanName,
		Tags:         binding.Status.Tags,
		Credentials:  make(map[string]json.RawMessage),
	}

	for sn, sd := range credentialsSecret.Data {
		vs.Credentials[sn] = sd
	}

	if binding.Status.VolumeStatus != nil {
		var mode string
		if binding.Status.VolumeStatus.ReadOnly {
			mode = "ro"
		} else {
			mode = "rw"
		}
		volumeMounts := []VolumeMount{
			{
				ContainerDir: binding.Status.VolumeStatus.Mount,
				DeviceType:   "shared",
				Mode:         mode,
			},
		}

		vs.VolumeMounts = volumeMounts
	}

	return vs
}

// getBindingName returns a pointer to the binding name set by the user, or nil if the name was not set.
func getBindingName(binding kfv1alpha1.ServiceInstanceBinding) *string {
	bindingOverride := binding.Spec.BindingNameOverride
	if bindingOverride != "" {
		return &bindingOverride
	}
	return nil
}
