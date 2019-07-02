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
	"github.com/google/kf/pkg/kf/internal/envutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KfApp provides a facade around Knative services for accessing and mutating its
// values.
type KfApp serving.Service

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

func (k *KfApp) getOrCreateRevisionTemplateSpec() *serving.RevisionTemplateSpec {
	if k.Spec.Template == nil {
		k.Spec.Template = &serving.RevisionTemplateSpec{}
	}

	return k.Spec.Template
}

func (k *KfApp) getRevisionTemplateSpecOrNil() *serving.RevisionTemplateSpec {
	if k == nil {
		return nil
	}
	return k.Spec.Template
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

// ToService casts this alias back into a Service.
func (k *KfApp) ToService() *serving.Service {
	svc := serving.Service(*k)
	return &svc
}

// NewKfApp creates a new KfApp.
func NewKfApp() KfApp {
	return KfApp{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}
}

// NewFromService creates a new KfApp from the given service pointer
// modifications to the KfApp will affect the underling svc.
func NewFromService(svc *serving.Service) *KfApp {
	return (*KfApp)(svc)
}
