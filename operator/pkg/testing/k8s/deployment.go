// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	mfTesting "kf-operator/pkg/testing/manifestival"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// DeploymentOption enables further configuration of a Deployment.
type DeploymentOption func(*apps.Deployment)

// ManifestivalDeployment creates a Deployment with name and options owned by manifestival.
func ManifestivalDeployment(name string, options ...DeploymentOption) *apps.Deployment {
	obj := Deployment(name, options...)
	mfTesting.SetManifestivalAnnotation(obj)
	mfTesting.SetLastApplied(obj)
	return obj
}

// Deployment creates a Deployment with Name name and Namespace test,
// and then applies DeploymentOptions to it.
func Deployment(name string, do ...DeploymentOption) *apps.Deployment {
	dep := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range do {
		opt(dep)
	}
	return dep
}

// WithDeploymentOwnerRefs creates a DeploymentOption that updates OwnerReferences
// in Deployment.
func WithDeploymentOwnerRefs(refs ...*metav1.OwnerReference) DeploymentOption {
	return func(dep *apps.Deployment) {
		references := dep.GetOwnerReferences()
		for _, ref := range refs {
			references = append(references, *ref)
		}
		dep.SetOwnerReferences(references)
	}
}

// WithDeploymentNamespace creates a DeploymentOption that sets Namespace in Deployment.
func WithDeploymentNamespace(namespace string) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.SetNamespace(namespace)
	}
}

// WithDeploymentAnnotation creates a DeploymentOption that sets annotation in Deployment.
func WithDeploymentAnnotation(annotations map[string]string) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.SetAnnotations(annotations)
	}
}

// WithDeploymentContainer creates a DeploymentOption that updates Containers
// in Deployment.
func WithDeploymentContainer(c *corev1.Container) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, *c)
	}
}

// WithDeploymentImagePullSecrets creates a DeploymentOption that updates ImagePullSecrets
// in Deployment.
func WithDeploymentImagePullSecrets(c corev1.LocalObjectReference) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.Spec.Template.Spec.ImagePullSecrets = append(dep.Spec.Template.Spec.ImagePullSecrets, c)
	}
}

// WithDeploymentInitContainer creates a DeploymentOption that updates InitContainers
// in Deployment.
func WithDeploymentInitContainer(c *corev1.Container) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.Spec.Template.Spec.InitContainers = append(dep.Spec.Template.Spec.Containers, *c)
	}
}

// WithDeploymentVolumes creates a DeploymentOption that updates Volumes
// in Deployment.
func WithDeploymentVolumes(v *corev1.Volume) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, *v)
	}
}

// WithDeploymentCreationTimestamp creates a DeploymentOption that sets
// CreationTimestamp in Deployment.
func WithDeploymentCreationTimestamp(t metav1.Time) DeploymentOption {
	return func(dep *apps.Deployment) {
		dep.ObjectMeta.CreationTimestamp = t
	}
}

// DeleteDeploymentAction creates a DeleteActionImpl that deletes
// deployments in Namespace test.
func DeleteDeploymentAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: "test",
			Verb:      "delete",
			Resource: schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
		},
		Name: name,
	}
}

// DeploymentReady marks the deployment as ready.
func DeploymentReady(dep *apps.Deployment) {
	readyCondition := []apps.DeploymentCondition{
		{
			Type:   apps.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		},
	}
	dep.Status = apps.DeploymentStatus{
		Conditions: readyCondition,
	}
}
