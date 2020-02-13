// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cfutil

import (
	"errors"
	"fmt"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	"github.com/google/kf/pkg/internal/envutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SystemEnvInjector is a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
type SystemEnvInjector interface {
	// GetVcapServices gets a VCAP_SERVICES compatible environment variable.
	GetVcapServices(appName string, bindings []servicecatalogv1beta1.ServiceBinding) ([]VcapService, error)

	// GetVcapService gets a VCAP_SERVICES service entry.
	GetVcapService(appName string, binding *servicecatalogv1beta1.ServiceBinding) (VcapService, error)

	// ComputeSystemEnv computes the environment variables that should be injected
	// on a given service.
	ComputeSystemEnv(app *v1alpha1.App, serviceBindings []servicecatalogv1beta1.ServiceBinding) (computed []corev1.EnvVar, err error)

	// GetClassFromInstance gets the service class for the given instance.
	GetClassFromInstance(instance *servicecatalogv1beta1.ServiceInstance) (*servicecatalogv1beta1.CommonServiceClassSpec, error)
}

type systemEnvInjector struct {
	client    servicecatalogclient.Interface
	k8sclient kubernetes.Interface
}

// NewSystemEnvInjector creates a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
func NewSystemEnvInjector(
	client servicecatalogclient.Interface,
	k8sclient kubernetes.Interface) SystemEnvInjector {
	return &systemEnvInjector{
		client:    client,
		k8sclient: k8sclient,
	}
}

func GetVcapServicesMap(appName string, services []VcapService) (VcapServicesMap, error) {
	out := VcapServicesMap{}
	for _, service := range services {
		out.Add(service)
	}
	return out, nil
}

func (s *systemEnvInjector) GetVcapService(appName string, binding *servicecatalogv1beta1.ServiceBinding) (VcapService, error) {

	secret, err := s.k8sclient.
		CoreV1().
		Secrets(binding.Namespace).
		Get(binding.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		return VcapService{}, fmt.Errorf("couldn't create VCAP_SERVICES, the secret for binding %s couldn't be fetched: %v", binding.Name, err)
	}

	serviceInstance, err := s.client.
		ServicecatalogV1beta1().
		ServiceInstances(binding.Namespace).
		Get(binding.Spec.InstanceRef.Name, metav1.GetOptions{})
	if err != nil {
		return VcapService{}, fmt.Errorf("couldn't get instance: %v", err)
	}

	class, err := s.GetClassFromInstance(serviceInstance)
	if err != nil {
		return VcapService{}, fmt.Errorf("couldn't get instance: %v", err)
	}

	return NewVcapService(*class, *serviceInstance, *binding, secret), nil
}

// GetClassFromInstance gets the service class for the given instance.
func (s *systemEnvInjector) GetClassFromInstance(instance *servicecatalogv1beta1.ServiceInstance) (*servicecatalogv1beta1.CommonServiceClassSpec, error) {
	if ref := instance.Spec.ClusterServiceClassRef; ref != nil {
		plan, err := s.client.
			ServicecatalogV1beta1().
			ClusterServiceClasses().
			Get(ref.Name, metav1.GetOptions{})

		if err != nil {
			return nil, err
		}

		return &plan.Spec.CommonServiceClassSpec, nil
	}

	if ref := instance.Spec.ServiceClassRef; ref != nil {
		plan, err := s.client.
			ServicecatalogV1beta1().
			ServiceClasses(instance.Namespace).
			Get(ref.Name, metav1.GetOptions{})

		if err != nil {
			return nil, err
		}

		return &plan.Spec.CommonServiceClassSpec, nil
	}

	return nil, errors.New("neither ClusterServiceClassRef nor ServiceClassRef were provided")
}

func (s *systemEnvInjector) GetVcapServices(appName string, bindings []servicecatalogv1beta1.ServiceBinding) (services []VcapService, err error) {
	for _, binding := range bindings {
		service, err := s.GetVcapService(appName, &binding)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *systemEnvInjector) ComputeSystemEnv(app *v1alpha1.App, serviceBindings []servicecatalogv1beta1.ServiceBinding) (computed []corev1.EnvVar, err error) {
	va, err := CreateVcapApplication(app)
	if err != nil {
		return nil, err
	}
	computed = append(computed, va)

	services, err := s.GetVcapServices(app.Name, serviceBindings)
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
