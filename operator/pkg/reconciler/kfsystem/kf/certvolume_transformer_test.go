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

package kf

import (
	"context"
	"testing"

	"kf-operator/pkg/testing/k8s"
	mftesting "kf-operator/pkg/testing/manifestival"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCertVolumeTransformer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     mftesting.Object
		secret *corev1.Secret
		want   *unstructured.Unstructured
	}{
		{
			name: "not a deployment",
			in:   k8s.ConfigMap("test-configmap"),
			want: mftesting.ToUnstructured(k8s.ConfigMap("test-configmap")),
		},
		{
			name: "not the controller",
			in:   k8s.Deployment("deployment-something"),
			want: mftesting.ToUnstructured(k8s.Deployment("deployment-something")),
		},
		{
			name: "volumes from the secret and manifest are added",
			in: k8s.Deployment(
				"controller",
				k8s.WithDeploymentVolumes(k8s.Volume("some-volume")),
				k8s.WithDeploymentContainer(k8s.Container(
					"some-container", "some-image",
					k8s.WithContainerVolumeMounts([]corev1.VolumeMount{
						{Name: "some-mount"},
					}),
				)),
			),
			want: mftesting.ToUnstructured(k8s.Deployment(
				"controller",
				k8s.WithDeploymentVolumes(k8s.Volume("some-volume")),
				k8s.WithDeploymentVolumes(k8s.Volume(
					"some-secret",
					k8s.WithVolumeSecretSource(&corev1.SecretVolumeSource{
						SecretName: "some-secret",
						Items: []corev1.KeyToPath{
							{Key: "foo.pem", Path: "foo.pem"},
							{Key: "bar.pem", Path: "bar.pem"},
						},
					}),
				)),
				k8s.WithDeploymentContainer(k8s.Container(
					"some-container", "some-image",
					k8s.WithContainerVolumeMounts([]corev1.VolumeMount{
						{Name: "some-mount"},
						{
							Name:      "some-secret",
							MountPath: "/etc/ssl/certs/foo.pem",
							SubPath:   "foo.pem",
							ReadOnly:  true,
						},
						{
							Name:      "some-secret",
							MountPath: "/etc/ssl/certs/bar.pem",
							SubPath:   "bar.pem",
							ReadOnly:  true,
						},
					}),
				)),
			)),
			secret: k8s.Secret("some-secret", k8s.WithSecretData(map[string][]byte{
				"foo.pem": []byte("some-cert"),
				"bar.pem": []byte("some-other-cert"),
			})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mftesting.ToUnstructured(tt.in)

			transformer := CertVolumeTransformer(
				context.Background(),
				tt.secret,
			)

			err := transformer(u)
			if err != nil {
				t.Error("Got error", err)
			}

			if diff := cmp.Diff(tt.want, u); diff != "" {
				t.Errorf("(-want, +got) = %v", diff)
			}
		})
	}
}
