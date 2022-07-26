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

package transformations_test

import (
	"kf-operator/pkg/operand/transformations"
	mftesting "kf-operator/pkg/testing/manifestival"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"

	. "kf-operator/pkg/testing/k8s"

	corev1 "k8s.io/api/core/v1"
)

func TestAppendDockerCredentials(t *testing.T) {
	targetDeploymentName := "deployment"
	targetSecretName := "secret"

	tests := []struct {
		name      string
		object    runtime.Object
		want      runtime.Object
		wantError bool
	}{{
		name:   "name doesn't match, no transform",
		object: Deployment("no-tranform"),
		want:   Deployment("no-tranform"),
	}, {
		name:   "not deployment, no transform",
		object: ConfigMap("config-map"),
		want:   ConfigMap("config-map"),
	}, {
		name: "append docker credentials",
		object: Deployment(targetDeploymentName, WithDeploymentContainer(
			&corev1.Container{
				Name: "container",
			},
		)),
		want: Deployment(targetDeploymentName,
			WithDeploymentContainer(
				&corev1.Container{
					Name: "container",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "docker",
							MountPath: "/etc/docker",
							ReadOnly:  true,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "DOCKER_CONFIG",
							Value: "/etc/docker",
						},
					},
				},
			),
			WithDeploymentVolumes(
				&corev1.Volume{
					Name: "docker",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: targetSecretName,
							Items: []corev1.KeyToPath{
								{
									Key:  ".dockerconfigjson",
									Path: "config.json",
								},
							},
						},
					},
				},
			),
		),
	}}

	transformer := transformations.AppendDockerCredentials(targetDeploymentName, targetSecretName)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.object)
			err := transformer(u)
			if gotError := (err != nil); tt.wantError != gotError {
				t.Errorf("Expect error %t, but got %t", tt.wantError, gotError)
			}
			if err != nil {
				return
			}
			want := mftesting.ToUnstructured(tt.want)

			if diff := cmp.Diff(want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}

func TestAddWICheckForSubresourceAPI(t *testing.T) {
	tests := []struct {
		name      string
		object    runtime.Object
		want      runtime.Object
		wantError bool
	}{{
		name:   "name doesn't match, no transform",
		object: Deployment("no-tranform"),
		want:   Deployment("no-tranform"),
	}, {
		name:   "not deployment, no transform",
		object: ConfigMap("config-map"),
		want:   ConfigMap("config-map"),
	}, {
		name:   "Init container added",
		object: Deployment(transformations.SubresourceAPIDeploymentName),
		want: Deployment(transformations.SubresourceAPIDeploymentName,
			WithDeploymentInitContainer(
				&corev1.Container{
					Name:    "wait-for-wi",
					Image:   "gcr.io/google.com/cloudsdktool/cloud-sdk:326.0.0-alpine",
					Command: []string{"bash"},
					Args: []string{
						"-c",
						"curl -s -H 'Metadata-Flavor: Google' 'http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token' --retry 30 --retry-connrefused --retry-max-time 30 > /dev/null || exit 1",
					},
				},
			),
		),
	}}

	transformer := transformations.AddWICheckForSubresourceAPI()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.object)
			err := transformer(u)
			if gotError := (err != nil); tt.wantError != gotError {
				t.Errorf("Expect error %t, but got %t", tt.wantError, gotError)
			}
			if err != nil {
				return
			}
			want := mftesting.ToUnstructured(tt.want)

			if diff := cmp.Diff(want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}
