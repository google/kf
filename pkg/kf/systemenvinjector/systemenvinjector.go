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

package systemenvinjector

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/internal/cfutil"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	corev1 "k8s.io/api/core/v1"
)

// SystemEnvInjectorInterface is the interface to interact with SystemEnvInjector
// through.
type SystemEnvInjectorInterface interface {
	ComputeSystemEnv(app *v1alpha1.App) (computed []corev1.EnvVar, err error)
}

// NewSystemEnvInjector creates a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
func NewSystemEnvInjector(bindingsClient servicebindings.ClientInterface) SystemEnvInjectorInterface {
	return &SystemEnvInjector{
		bindingsClient: bindingsClient,
	}
}

// SystemEnvInjector is a utility used to update v1alpha1.Apps with
// CF style system environment variables like VCAP_SERVICES.
type SystemEnvInjector struct {
	bindingsClient servicebindings.ClientInterface
}

// ComputeSystemEnv computes the environment variables that should be injected
// on a given service.
func (s *SystemEnvInjector) ComputeSystemEnv(app *v1alpha1.App) (computed []corev1.EnvVar, err error) {
	va, err := cfutil.CreateVcapApplication(app)
	if err != nil {
		return nil, err
	}
	computed = append(computed, va)

	vs, err := s.bindingsClient.GetVcapServices(app.Name, servicebindings.WithGetVcapServicesNamespace(app.Namespace))
	if err != nil {
		return nil, err
	}
	vsVar, err := envutil.NewJSONEnvVar("VCAP_SERVICES", vs)
	if err != nil {
		return nil, err
	}
	computed = append(computed, vsVar)

	return
}
