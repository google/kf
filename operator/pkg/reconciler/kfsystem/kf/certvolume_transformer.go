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
	"path/filepath"

	mf "github.com/manifestival/manifestival"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/logging"
)

// CertVolumeTransformer adds the necessary volume and volumeMounts for the
// given secret.
func CertVolumeTransformer(ctx context.Context, secret *corev1.Secret) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "Deployment" || u.GetName() != "controller" {
			return nil
		}

		log := logging.FromContext(ctx)
		d := new(apps.Deployment)
		if err := scheme.Scheme.Convert(u, d, nil); err != nil {
			log.Error(err, "Error converting Unstructured to Deployment", "unstructured", u, "deployment", d)
			return err
		}

		var (
			keys   []corev1.KeyToPath
			mounts []corev1.VolumeMount
		)

		for certFile := range secret.Data {
			keys = append(keys, corev1.KeyToPath{
				Key:  certFile,
				Path: certFile,
			})
			mounts = append(mounts, corev1.VolumeMount{
				Name:      secret.Name,
				MountPath: filepath.Join("/etc/ssl/certs", certFile),
				SubPath:   certFile,
				ReadOnly:  true,
			})
		}

		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: secret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name,
					Items:      keys,
				},
			},
		})

		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(
			d.Spec.Template.Spec.Containers[0].VolumeMounts,
			mounts...,
		)

		return scheme.Scheme.Convert(d, u, nil)
	}
}
