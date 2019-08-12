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
	"fmt"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	clientv1beta1 "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/typed/servicecatalog/v1beta1"
	"github.com/google/kf/pkg/internal/envutil"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// BindingNameLabel is the label used on bindings to define what VCAP name the secret should be rooted under.
	BindingNameLabel = "kf-binding-name"
	// AppNameLabel is the label used on bindings to define which app the binding belongs to.
	AppNameLabel = "kf-app-name"
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
		BindingName:  binding.Labels[BindingNameLabel],
		Name:         binding.Name,
		InstanceName: binding.Spec.InstanceRef.Name,
		Label:        instance.Spec.ClusterServiceClassExternalName,
		Plan:         instance.Spec.ClusterServicePlanExternalName,
		Credentials:  make(map[string]string),
	}

	// Make sure we can work with both ServiceClass and ClusterServiceClass
	if instance.Spec.ServiceClassExternalName != "" {
		vs.Label = instance.Spec.ServiceClassExternalName
	}

	if instance.Spec.ServicePlanExternalName != "" {
		vs.Plan = instance.Spec.ServicePlanExternalName
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

func GetVcapServicesMap(appName string, services []VcapService) (VcapServicesMap, error) {
	out := VcapServicesMap{}
	for _, service := range services {
		out.Add(service)
	}
	return out, nil
}

func GetVcapService(appName string, binding apiv1beta1.ServiceBinding, client clientv1beta1.ServiceInstanceInterface, secretsClient clientcorev1.SecretInterface) (VcapService, error) {

	secret, err := secretsClient.Get(binding.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		return VcapService{}, fmt.Errorf("couldn't create VCAP_SERVICES, the secret for binding %s couldn't be fetched: %v", binding.Name, err)
	}

	serviceInstance, err := client.Get(binding.Spec.InstanceRef.Name, metav1.GetOptions{})
	if err != nil {
		return VcapService{}, nil
	}

	return NewVcapService(*serviceInstance, binding, secret), nil
}

func GetVcapServices(appName string, bindings []apiv1beta1.ServiceBinding, client clientv1beta1.ServiceInstanceInterface, secretsClient clientcorev1.SecretInterface) ([]VcapService, error) {
	var services []VcapService
	for _, binding := range bindings {
		service, err := GetVcapService(appName, binding, client, secretsClient)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func ComputeSystemEnv(app *v1alpha1.App, serviceBindings []apiv1beta1.ServiceBinding, client clientv1beta1.ServiceInstanceInterface, secretsClient clientcorev1.SecretInterface) (computed []corev1.EnvVar, err error) {
	va, err := CreateVcapApplication(app)
	if err != nil {
		return nil, err
	}
	computed = append(computed, va)

	services, err := GetVcapServices(app.Name, serviceBindings, client, secretsClient)
	if err != nil {
		return nil, err
	}

	serviceMap, err := GetVcapServicesMap(app.Name, services)
	if err != nil {
		return nil, err
	}

	vsVar, err := envutil.NewJSONEnvVar("VCAP_SERVICES", serviceMap)
	if err != nil {
		return nil, err
	}
	computed = append(computed, vsVar)

	return
}
