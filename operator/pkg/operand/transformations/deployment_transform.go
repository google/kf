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

package transformations

import (
	mf "github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"
)

const (
	// SubresourceAPIDeploymentName is the name of the Subresource API deployment.
	SubresourceAPIDeploymentName = "subresource-apiserver"
)

// AddWICheckForSubresourceAPI add an init-container to check for WI availability.
func AddWICheckForSubresourceAPI() mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "Deployment" || u.GetName() != SubresourceAPIDeploymentName {
			return nil
		}

		deployment := &appsv1.Deployment{}
		if err := scheme.Scheme.Convert(u, deployment, nil); err != nil {
			return err
		}

		deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers,
			corev1.Container{
				Name:    "wait-for-wi",
				Image:   "gcr.io/google.com/cloudsdktool/cloud-sdk:326.0.0-alpine",
				Command: []string{"bash"},
				Args: []string{
					"-c",
					"curl -s -H 'Metadata-Flavor: Google' 'http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token' --retry 30 --retry-connrefused --retry-max-time 30 > /dev/null || exit 1",
				},
			},
		)

		return scheme.Scheme.Convert(deployment, u, nil)
	}

}

// AppendDockerCredentials append docker credentials to deployments matching the name.
func AppendDockerCredentials(deploymentName string, secretName string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "Deployment" || u.GetName() != deploymentName {
			return nil
		}

		deployment := &appsv1.Deployment{}
		if err := scheme.Scheme.Convert(u, deployment, nil); err != nil {
			return err
		}

		containers := deployment.Spec.Template.Spec.Containers
		if len(containers) == 0 {
			return nil
		}

		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "docker",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
					Items: []corev1.KeyToPath{
						{
							Key:  ".dockerconfigjson",
							Path: "config.json",
						},
					},
				},
			},
		})

		containers[0].VolumeMounts = append(containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "docker",
			MountPath: "/etc/docker",
			ReadOnly:  true,
		})

		containers[0].Env = append(containers[0].Env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/etc/docker",
		})

		return scheme.Scheme.Convert(deployment, u, nil)
	}
}
