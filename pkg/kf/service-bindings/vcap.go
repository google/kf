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
