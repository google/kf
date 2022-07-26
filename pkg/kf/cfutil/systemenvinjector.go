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
	"context"
	"encoding/json"
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// DatabaseURLEnvVarName is the environment variable expected by
	// applications looking for a CF style database URI.
	DatabaseURLEnvVarName = "DATABASE_URL"

	// VcapServicesEnvVarName is the environment variable expected by
	// applications looking CF style service injection details.
	VcapServicesEnvVarName = "VCAP_SERVICES"
)

// SystemEnvInjector is a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
type SystemEnvInjector interface {
	// GetVcapServices gets a VCAP_SERVICES compatible environment variable.
	GetVcapServices(ctx context.Context, appName string, bindings []v1alpha1.ServiceInstanceBinding) ([]VcapService, error)

	// GetVcapService gets a VCAP_SERVICES service entry.
	GetVcapService(ctx context.Context, appName string, binding v1alpha1.ServiceInstanceBinding) (VcapService, error)

	// ComputeSystemEnv computes the environment variables that should be injected
	// on a given service.
	ComputeSystemEnv(ctx context.Context, app *v1alpha1.App, serviceBindings []v1alpha1.ServiceInstanceBinding) (computed []corev1.EnvVar, err error)
}

type systemEnvInjector struct {
	k8sclient kubernetes.Interface
}

// NewSystemEnvInjector creates a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
func NewSystemEnvInjector(k8sclient kubernetes.Interface) SystemEnvInjector {
	return &systemEnvInjector{
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

// GetVcapService gets a VCAP_SERVICES entry for a service instance binding.
func (s *systemEnvInjector) GetVcapService(ctx context.Context, appName string, binding v1alpha1.ServiceInstanceBinding) (VcapService, error) {
	secret, err := s.k8sclient.
		CoreV1().
		Secrets(binding.Namespace).
		Get(ctx, binding.Status.CredentialsSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return VcapService{}, fmt.Errorf("couldn't create VCAP_SERVICES, the secret for binding %s couldn't be fetched: %v", binding.Name, err)
	}

	return NewVcapService(binding, *secret), nil
}

func (s *systemEnvInjector) GetVcapServices(ctx context.Context, appName string, bindings []v1alpha1.ServiceInstanceBinding) (services []VcapService, err error) {
	for _, binding := range bindings {
		service, err := s.GetVcapService(ctx, appName, binding)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

// ComputeSystemEnv gets the env vars for an app and its associated service bindings.
func (s *systemEnvInjector) ComputeSystemEnv(ctx context.Context, app *v1alpha1.App, serviceBindings []v1alpha1.ServiceInstanceBinding) (computed []corev1.EnvVar, err error) {
	services, err := s.GetVcapServices(ctx, app.Name, serviceBindings)
	if err != nil {
		return nil, err
	}

	// Set DATABASE_URL
	// https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#DATABASE-URL
	for _, svc := range services {
		if uri, ok := svc.Credentials["uri"]; ok {
			// Cloud Foundry doesn't check whether the value at the "uri" key is a valid string, but we check that the
			// value can be unmarshaled into a string. Unmarshaling returns the desired uri string rather than the JSON encoded string.
			// https://github.com/cloudfoundry/cloud_controller_ng/blob/25bd86e7ddcd12ca045836ce648ae48af4f07534/lib/cloud_controller/database_uri_generator.rb
			var uriStr string
			err := json.Unmarshal(uri, &uriStr)
			if err == nil {
				computed = append(computed, corev1.EnvVar{
					Name:  DatabaseURLEnvVarName,
					Value: uriStr,
				})
				break
			}
		}
	}

	serviceMap, err := GetVcapServicesMap(app.Name, services)
	if err != nil {
		return nil, err
	}

	vsVar, err := envutil.NewJSONEnvVar(VcapServicesEnvVarName, serviceMap)
	if err != nil {
		return nil, err
	}
	computed = append(computed, vsVar)

	return
}
