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

package manifest

import (
	"github.com/docker/distribution/reference"
	"k8s.io/client-go/kubernetes/scheme"

	mf "github.com/manifestival/manifestival"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// VersionFromLabel looks through the manifest for a resource with the label and returns the label value
func VersionFromLabel(manifest *mf.Manifest, label, defaultValue string) string {
	for _, resource := range manifest.Resources() {
		labels := resource.GetLabels()
		if labels != nil && labels[label] != "" {
			return labels[label]
		}
	}
	return defaultValue
}

// VersionFromDeploymentImage looks through the manifest for a resource with the label and returns the label value
func VersionFromDeploymentImage(manifest *mf.Manifest, defaultValue string) string {
	for _, u := range manifest.Resources() {
		if u.GetKind() == "Deployment" {
			var deployment = &appsv1.Deployment{}
			if err := scheme.Scheme.Convert(&u, deployment, nil); err != nil {
				continue
			}
			if imageVersion := imageVersionFromPodspec(&deployment.Spec.Template.Spec); imageVersion != "" {
				return imageVersion
			}
		}
	}
	return defaultValue
}

func imageVersionFromPodspec(spec *corev1.PodSpec) string {
	containers := spec.Containers
	for index := range containers {
		container := &containers[index]
		if imageTag := imageTag(container.Image); imageTag != "" {
			return imageTag
		}
	}
	return ""
}

func imageTag(image string) string {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return ""
	}

	nameTagged, ok := ref.(reference.NamedTagged)
	if !ok {
		return ""
	}
	return nameTagged.Tag()
}
