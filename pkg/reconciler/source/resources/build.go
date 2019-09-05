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
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/serving/pkg/resources"
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

func makeContainerImageBuild(source *v1alpha1.Source) (*build.Build, error) {
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
	}, nil
}

func makeDockerImageBuild(source *v1alpha1.Source) (*build.Build, error) {
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
	}, nil
}

func makeBuildpackBuild(source *v1alpha1.Source) (*build.Build, error) {
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
				Env: source.Spec.BuildpackBuild.Env,
			},
		},
	}, nil
}

func makeObjectMeta(source *v1alpha1.Source) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      BuildName(source),
		Namespace: source.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(source),
		},
		// Copy labels from the parent
		Labels: resources.UnionMaps(
			source.GetLabels(), map[string]string{
				managedByLabel: "kf",
			}),
	}
}

// MakeBuild creates a Build for a Source.
func MakeBuild(source *v1alpha1.Source) (*build.Build, error) {
	switch {
	case source.Spec.IsContainerBuild():
		return makeContainerImageBuild(source)
	case source.Spec.IsDockerfileBuild():
		return makeDockerImageBuild(source)
	default:
		return makeBuildpackBuild(source)
	}
}
