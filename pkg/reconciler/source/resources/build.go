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
	"fmt"

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
)

// BuildName gets the name of a Build for a Source.
func BuildName(source *v1alpha1.Source) string {
	return fmt.Sprintf("%s-%d", source.Name, source.Generation)
}

// AppImageName gets the image name for an application.
func AppImageName(source *v1alpha1.Source) string {
	return fmt.Sprintf("app-%s-%s:%d", source.Namespace, source.Name, source.Generation)
}

// JoinRepositoryImage joins a repository and image name.
func JoinRepositoryImage(repository, imageName string) string {
	return fmt.Sprintf("%s/%s", repository, imageName)
}

func makeContainerImageBuild(source *v1alpha1.Source) (*build.Build, error) {
	buildName := BuildName(source)

	args := []build.ArgumentSpec{
		{
			Name:  v1alpha1.BuildArgImage,
			Value: source.Spec.ContainerImage.Image,
		},
	}

	return &build.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName,
			Namespace: source.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(source),
			},
			// Copy labels from the parent
			Labels: resources.UnionMaps(
				source.GetLabels(), map[string]string{
					managedByLabel: "kf",
				}),
		},
		Spec: build.BuildSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			Template: &build.TemplateInstantiationSpec{
				Name:      containerImageTemplate,
				Kind:      "ClusterBuildTemplate",
				Arguments: args,
			},
		},
	}, nil
}

func makeBuildpackBuild(source *v1alpha1.Source) (*build.Build, error) {
	buildName := BuildName(source)
	appImageName := AppImageName(source)
	imageDestination := JoinRepositoryImage(source.Spec.BuildpackBuild.Registry, appImageName)

	buildSource := &build.SourceSpec{
		Custom: &corev1.Container{
			Image: source.Spec.BuildpackBuild.Source,
		},
	}

	args := []build.ArgumentSpec{
		{
			Name:  v1alpha1.BuildArgImage,
			Value: imageDestination,
		},
		{
			Name:  v1alpha1.BuildArgBuildpackBuilder,
			Value: source.Spec.BuildpackBuild.BuildpackBuilder,
		},
		{
			Name:  v1alpha1.BuildArgBuildpack,
			Value: source.Spec.BuildpackBuild.Buildpack,
		},
	}

	return &build.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName,
			Namespace: source.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(source),
			},
			// Copy labels from the parent
			Labels: resources.UnionMaps(
				source.GetLabels(), map[string]string{
					managedByLabel: "kf",
				}),
		},
		Spec: build.BuildSpec{
			Source:             buildSource,
			ServiceAccountName: source.Spec.ServiceAccount,
			Template: &build.TemplateInstantiationSpec{
				Name:      buildpackBuildTemplate,
				Kind:      "ClusterBuildTemplate",
				Arguments: args,
				Env:       source.Spec.BuildpackBuild.Env,
			},
		},
	}, nil
}

// MakeBuild creates a Build for a Source.
func MakeBuild(source *v1alpha1.Source) (*build.Build, error) {
	if source.Spec.IsContainerBuild() {
		return makeContainerImageBuild(source)
	} else {
		return makeBuildpackBuild(source)
	}
}
