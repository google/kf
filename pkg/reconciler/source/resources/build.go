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

package resources

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	build "github.com/google/kf/third_party/knative-build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const (
	managedByLabel         = "app.kubernetes.io/managed-by"
	buildpackBuildTemplate = "buildpack"
	containerImageTemplate = "container"
	dockerImageTemplate    = "kaniko"
)

// BuildName gets the name of a Build for a Source.
func BuildName(source *v1alpha1.Source) string {
	return source.Name
}

// BuildSecretName gets the name of a Secret for a Source.
func BuildSecretName(source *v1alpha1.Source) string {
	return BuildName(source)
}

func makeContainerImageBuild(source *v1alpha1.Source) (*build.Build, *corev1.Secret, error) {
	return &build.Build{
		ObjectMeta: makeObjectMeta(source),
		Spec: build.BuildSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			Template: &build.TemplateInstantiationSpec{
				Name: containerImageTemplate,
				Kind: "ClusterBuildTemplate",
				Arguments: []build.ArgumentSpec{
					{
						Name:  v1alpha1.BuildArgImage,
						Value: source.Spec.ContainerImage.Image,
					},
				},
			},
		},
	}, nil, nil
}

func makeDockerImageBuild(source *v1alpha1.Source) (*build.Build, *corev1.Secret, error) {
	return &build.Build{
		ObjectMeta: makeObjectMeta(source),
		Spec: build.BuildSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			Source: &build.SourceSpec{
				Custom: &corev1.Container{
					Image: source.Spec.Dockerfile.Source,
				},
			},
			Template: &build.TemplateInstantiationSpec{
				Name: dockerImageTemplate,
				Kind: "ClusterBuildTemplate",
				Arguments: []build.ArgumentSpec{
					{Name: v1alpha1.BuildArgImage, Value: source.Spec.Dockerfile.Image},
					{Name: v1alpha1.BuildArgDockerfile, Value: source.Spec.Dockerfile.Path},
				},
			},
		},
	}, nil, nil
}

func makeBuildpackBuild(source *v1alpha1.Source) (*build.Build, *corev1.Secret, error) {
	// We want to use a secret to store these, so we'll have to point at the
	// secret.
	env := []corev1.EnvVar{}
	for _, e := range source.Spec.BuildpackBuild.Env {
		env = append(env, corev1.EnvVar{
			Name: e.Name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: BuildSecretName(source),
					},
					Key: e.Name,
				},
			},
		})
	}

	return &build.Build{
		ObjectMeta: makeObjectMeta(source),
		Spec: build.BuildSpec{
			Source: &build.SourceSpec{
				Custom: &corev1.Container{
					Image: source.Spec.BuildpackBuild.Source,
				},
			},
			ServiceAccountName: source.Spec.ServiceAccount,
			Template: &build.TemplateInstantiationSpec{
				Name: buildpackBuildTemplate,
				Kind: "ClusterBuildTemplate",
				Arguments: []build.ArgumentSpec{
					{
						Name:  v1alpha1.BuildArgImage,
						Value: source.Spec.BuildpackBuild.Image,
					},
					{
						Name:  v1alpha1.BuildArgBuildpackBuilder,
						Value: source.Spec.BuildpackBuild.BuildpackBuilder,
					},
					{
						Name:  v1alpha1.BuildArgBuildpack,
						Value: source.Spec.BuildpackBuild.Buildpack,
					},
					{
						Name:  v1alpha1.BuildArgBuildpackRunImage,
						Value: source.Spec.BuildpackBuild.Stack,
					},
				},
				Env: env,
			},
		},
	}, makeSecret(source), nil
}

func makeSecret(source *v1alpha1.Source) *corev1.Secret {
	m := map[string][]byte{}

	// TODO(#821): Support source.Env.ValueFrom
	for _, e := range source.Spec.BuildpackBuild.Env {
		m[e.Name] = []byte(e.Value)
	}

	return &corev1.Secret{
		ObjectMeta: makeObjectMeta(source),
		Data:       m,
	}
}

func makeObjectMeta(source *v1alpha1.Source) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      BuildName(source),
		Namespace: source.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(source),
		},
		// Copy labels from the parent
		Labels: v1alpha1.UnionMaps(
			source.GetLabels(), map[string]string{
				managedByLabel: "kf",
			}),
	}
}

// MakeBuild creates a Build and Secret for a Source. The Secret CAN be nil if
// the Source does not require it.
func MakeBuild(source *v1alpha1.Source) (*build.Build, *corev1.Secret, error) {
	switch {
	case source.Spec.IsContainerBuild():
		return makeContainerImageBuild(source)
	case source.Spec.IsDockerfileBuild():
		return makeDockerImageBuild(source)
	default:
		return makeBuildpackBuild(source)
	}
}
