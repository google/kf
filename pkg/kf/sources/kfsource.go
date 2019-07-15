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

package sources

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// KfSource provides a facade around v1alpha1.Source for accessing and mutating
// its values.
type KfSource v1alpha1.Source

// GetName retrieves the name of the space.
func (k *KfSource) GetName() string {
	return k.Name
}

// SetName sets the name of the space.
func (k *KfSource) SetName(name string) {
	k.Name = name
}

// GetNamespace retrieves the namespace for the source.
func (k *KfSource) GetNamespace() string {
	return k.Namespace
}

// SetNamespace sets the namespace for the source.
func (k *KfSource) SetNamespace(namespace string) {
	k.Namespace = namespace
}

// SetContainerImageSource sets an image as a container image source.
func (k *KfSource) SetContainerImageSource(sourceImage string) {
	k.Spec.ContainerImage.Image = sourceImage
}

// GetContainerImageSource gets the container image source.
func (k *KfSource) GetContainerImageSource() string {
	return k.Spec.ContainerImage.Image
}

// GetBuildpackBuildSource returns the image that contins the build source if
// this is a buildpack style build.
func (k *KfSource) GetBuildpackBuildSource() string {
	return k.Spec.BuildpackBuild.Source
}

// SetBuildpackBuildSource sets the image that contains the source code.
func (k *KfSource) SetBuildpackBuildSource(sourceImage string) {
	k.Spec.BuildpackBuild.Source = sourceImage
}

// SetBuildpackBuildRegistry sets the container registry that the built code
// will be pushed to.
func (k *KfSource) SetBuildpackBuildRegistry(registry string) {
	k.Spec.BuildpackBuild.Registry = registry
}

// GetBuildpackBuildRegistry returns the container registry that the built code
// will be pushed to.
func (k *KfSource) GetBuildpackBuildRegistry() string {
	return k.Spec.BuildpackBuild.Registry
}

// SetBuildpackBuildEnv sets the environment variables for a buildpack build.
func (k *KfSource) SetBuildpackBuildEnv(env []corev1.EnvVar) {
	k.Spec.BuildpackBuild.Env = env
}

// GetBuildpackBuildEnv sets the environment variables for a buildpack build.
func (k *KfSource) GetBuildpackBuildEnv() []corev1.EnvVar {
	return k.Spec.BuildpackBuild.Env
}

// SetBuildpackBuildBuildpack sets the buildpack for a buildpack build.
func (k *KfSource) SetBuildpackBuildBuildpack(buildpack string) {
	k.Spec.BuildpackBuild.Buildpack = buildpack
}

// GetBuildpackBuildBuildpack gets the buildpack for a buildpack build.
func (k *KfSource) GetBuildpackBuildBuildpack() string {
	return k.Spec.BuildpackBuild.Buildpack
}

// ToSource casts this alias back into a Namespace.
func (k *KfSource) ToSource() *v1alpha1.Source {
	return (*v1alpha1.Source)(k)
}

// NewKfSource creates a new KfSource.
func NewKfSource() KfSource {
	return KfSource{}
}
