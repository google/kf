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

package apps

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"

	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/sources"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KfApp provides a facade around Knative services for accessing and mutating its
// values.
type KfApp v1alpha1.App

// GetName retrieves the name of the app.
func (k *KfApp) GetName() string {
	return k.Name
}

// SetName sets the name of the app.
func (k *KfApp) SetName(name string) {
	k.Name = name
}

// SetNamespace sets the namespace for the app.
func (k *KfApp) SetNamespace(namespace string) {
	k.Namespace = namespace
}

// GetNamespace gets the namespace for the app.
func (k *KfApp) GetNamespace() string {
	return k.Namespace
}

func (k *KfApp) getOrCreateRevisionTemplateSpec() *v1alpha1.AppSpecTemplate {
	return &k.Spec.Template
}

func (k *KfApp) getRevisionTemplateSpecOrNil() *v1alpha1.AppSpecTemplate {
	if k == nil {
		return nil
	}
	return &k.Spec.Template
}

func (k *KfApp) getOrCreateContainer() *corev1.Container {
	rl := k.getOrCreateRevisionTemplateSpec()
	if len(rl.Spec.Containers) == 0 {
		rl.Spec.Containers = []v1.Container{{}}
	}

	return &k.getOrCreateRevisionTemplateSpec().Spec.Containers[0]
}

func (k *KfApp) getContainerOrNil() *corev1.Container {
	if rl := k.getRevisionTemplateSpecOrNil(); rl != nil {
		if len(rl.Spec.Containers) != 0 {
			return &rl.Spec.Containers[0]
		}
	}

	return nil
}

// SetImage sets the image for the application and a policy to always refresh it.
func (k *KfApp) SetImage(imageName string) {
	container := k.getOrCreateContainer()
	container.ImagePullPolicy = "Always"
	container.Image = imageName
}

// GetImage gets the image associated with the container.
func (k *KfApp) GetImage() string {
	if container := k.getContainerOrNil(); container != nil {
		return container.Image
	}

	return ""
}

// SetContainerPorts sets the ports the container will open.
func (k *KfApp) SetContainerPorts(ports []corev1.ContainerPort) {
	k.getOrCreateContainer().Ports = ports
}

// GetContainerPorts gets the ports the container will open.
func (k *KfApp) GetContainerPorts() []corev1.ContainerPort {
	if container := k.getContainerOrNil(); container != nil {
		return container.Ports
	}

	return nil
}

// SetServiceAccount sets the account the application will run as.
func (k *KfApp) SetServiceAccount(sa string) {
	k.getOrCreateRevisionTemplateSpec().Spec.ServiceAccountName = sa
}

// SetSource sets the source the application will use to build.
func (k *KfApp) SetSource(src sources.KfSource) {
	k.Spec.Source = src.Spec
}

// GetServiceAccount returns the service account used by the container.
func (k *KfApp) GetServiceAccount() string {
	if rl := k.getRevisionTemplateSpecOrNil(); rl != nil {
		return rl.Spec.ServiceAccountName
	}

	return ""
}

// GetEnvVars reads the environment variables off an app.
func (k *KfApp) GetEnvVars() []corev1.EnvVar {
	if container := k.getContainerOrNil(); container != nil {
		return container.Env
	}

	return nil
}

// SetEnvVars sets environment variables on an app.
func (k *KfApp) SetEnvVars(env []corev1.EnvVar) {
	k.getOrCreateContainer().Env = env
}

// MergeEnvVars adds the environment variables listed to the existing ones,
// overwriting duplicates by key.
func (k *KfApp) MergeEnvVars(env []corev1.EnvVar) {
	k.SetEnvVars(envutil.DeduplicateEnvVars(append(k.GetEnvVars(), env...)))
}

// DeleteEnvVars removes environment variables with the given key.
func (k *KfApp) DeleteEnvVars(names []string) {
	k.SetEnvVars(envutil.RemoveEnvVars(names, k.GetEnvVars()))
}

// GetMemory gets memory request for the app.
func (k *KfApp) GetMemory() *resource.Quantity {
	if container := k.getContainerOrNil(); container != nil {
		if resourceRequests := container.Resources.Requests; resourceRequests != nil {
			memory, exists := resourceRequests[corev1.ResourceMemory]
			if exists {
				return &memory
			}
		}
	}
	return nil
}

// SetMemory sets memory request for the app.
func (k *KfApp) SetMemory(memory *resource.Quantity) {
	k.setResourceRequest(corev1.ResourceMemory, memory)
}

// GetStorage gets disk storage request for the app.
func (k *KfApp) GetStorage() *resource.Quantity {
	if container := k.getContainerOrNil(); container != nil {
		if resourceRequests := container.Resources.Requests; resourceRequests != nil {
			storage, exists := resourceRequests[corev1.ResourceEphemeralStorage]
			if exists {
				return &storage
			}
		}
	}
	return nil
}

// SetStorage sets disk storage request for the app.
func (k *KfApp) SetStorage(storage *resource.Quantity) {
	k.setResourceRequest(corev1.ResourceEphemeralStorage, storage)
}

// GetCPU gets CPU request for the app.
func (k *KfApp) GetCPU() *resource.Quantity {
	if container := k.getContainerOrNil(); container != nil {
		if resourceRequests := container.Resources.Requests; resourceRequests != nil {
			cpu, exists := resourceRequests[corev1.ResourceCPU]
			if exists {
				return &cpu
			}
		}
	}
	return nil
}

// SetCPU sets CPU request for the app.
func (k *KfApp) SetCPU(cpu *resource.Quantity) {
	k.setResourceRequest(corev1.ResourceCPU, cpu)
}

// Set a resource request for an app. Request amount can be cleared by passing in nil
func (k *KfApp) setResourceRequest(r v1.ResourceName, quantity *resource.Quantity) {
	container := k.getOrCreateContainer()
	resourceRequests := container.Resources.Requests

	if resourceRequests == nil {
		resourceRequests = v1.ResourceList{}
	}

	if quantity == nil {
		delete(resourceRequests, r)
	} else {
		resourceRequests[r] = *quantity
	}
	container.Resources.Requests = resourceRequests
}

// GetHealthCheck gets the readiness probe or nil if one doesn't exist.
func (k *KfApp) GetHealthCheck() *corev1.Probe {
	if cont := k.getContainerOrNil(); cont != nil {
		return cont.ReadinessProbe
	}

	return nil
}

// SetHealthCheck sets the readiness probe for the container.
func (k *KfApp) SetHealthCheck(probe *corev1.Probe) {
	container := k.getOrCreateContainer()
	container.ReadinessProbe = probe
}

func (k *KfApp) GetServiceBindings() []v1alpha1.AppSpecServiceBinding {
	return k.Spec.ServiceBindings
}

// ToApp casts this alias back into an App.
func (k *KfApp) ToApp() *v1alpha1.App {
	app := v1alpha1.App(*k)
	return &app
}

// NewKfApp creates a new KfApp.
func NewKfApp() KfApp {
	return KfApp{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "kf.dev/v1alpha1",
		},
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{}},
				},
			},
		},
	}
}

// NewFromApp creates a new KfApp from the given service pointer
// modifications to the KfApp will affect the underling app.
func NewFromApp(app *v1alpha1.App) *KfApp {
	return (*KfApp)(app)
}
