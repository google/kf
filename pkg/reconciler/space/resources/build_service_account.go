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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const imagePushSecretPrefix = "kf-registry"

// BuildServiceAccountName gets the name of the service account for given the
// space.
func BuildServiceAccountName(space *v1alpha1.Space) string {
	return space.Status.BuildConfig.ServiceAccount
}

// BuildImagePushSecretName generates a name for a secret given the original secret from config-secrets.
func BuildImagePushSecretName(secret *corev1.Secret) string {
	return v1alpha1.GenerateName(imagePushSecretPrefix, secret.Name)
}

// MakeBuildServiceAccount creates a ServiceAccount for build pipelines to
// use.
func MakeBuildServiceAccount(
	space *v1alpha1.Space,
	kfSecrets []*corev1.Secret,
	gsaName string,
	containerregistry string,
) (*corev1.ServiceAccount, []*corev1.Secret, error) {
	annotations := map[string]string{}
	if gsaName != "" {
		// Workload Identity
		annotations[v1alpha1.WorkloadIdentityAnnotation] = gsaName
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildServiceAccountName(space),
			Namespace: NamespaceName(space),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			Labels: v1alpha1.UnionMaps(space.GetLabels(), map[string]string{
				v1alpha1.ManagedByLabel: "kf",
			}),
			Annotations: annotations,
		},
	}

	// If WI is not enabled or no GSA is provided, then the build secrets set in config-secrets
	// should be referenced in the service account.
	var desiredSecrets []*corev1.Secret
	for idx, kfs := range kfSecrets {
		// We'll make a copy of the Data to ensure nothing gets altered anywhere.
		dataCopy := map[string][]byte{}
		for k, v := range kfs.Data {
			d := make([]byte, len(v))
			copy(d, v)
			dataCopy[k] = d
		}

		desiredSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      BuildImagePushSecretName(kfs),
				Namespace: NamespaceName(space),
				OwnerReferences: []metav1.OwnerReference{
					*kmeta.NewControllerRef(space),
				},
				Labels: v1alpha1.UnionMaps(space.GetLabels(), map[string]string{
					v1alpha1.ManagedByLabel: "kf",
				}),
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: dataCopy,
		}

		// Docker secrets needs to be properly annotation for Tekton to pick up.
		// https://github.com/tektoncd/pipeline/blob/main/docs/auth.md
		if containerregistry != "" {
			if desiredSecret.Annotations == nil {
				desiredSecret.Annotations = make(map[string]string)
			}
			desiredSecret.Annotations[fmt.Sprintf("tekton.dev/docker-%d", idx)] = containerregistry
		}

		desiredSecrets = append(desiredSecrets, desiredSecret)
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: desiredSecret.Name,
		})
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: desiredSecret.Name,
		})
	}

	return sa, desiredSecrets, nil
}
