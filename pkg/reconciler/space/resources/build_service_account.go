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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// BuildServiceAccountName gets the name of the service account for given the
// space.
func BuildServiceAccountName(space *v1alpha1.Space) string {
	return v1alpha1.DefaultBuildServiceAccountName
}

// BuildSecretName gets the name of the secret for given the space.
func BuildSecretName(space *v1alpha1.Space) string {
	// We'll just use the same name as the service account
	return BuildServiceAccountName(space)
}

// MakeBuildServiceAccount creates a ServiceAccount for build pipelines to
// use.
func MakeBuildServiceAccount(
	space *v1alpha1.Space,
	kfSecrets []*corev1.Secret,
) (*corev1.ServiceAccount, []*corev1.Secret, error) {
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
		},
		Secrets: []corev1.ObjectReference{
			{Name: BuildSecretName(space)},
		},
	}

	var desiredSecrets []*corev1.Secret
	for _, kfs := range kfSecrets {
		// We'll make a copy of the Data to ensure nothing gets altered anywhere.
		dataCopy := map[string][]byte{}
		for k, v := range kfs.Data {
			d := make([]byte, len(v))
			copy(d, v)
			dataCopy[k] = d
		}

		desiredSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      BuildSecretName(space),
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

		desiredSecrets = append(desiredSecrets, desiredSecret)
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: desiredSecret.Name,
		})
	}

	return sa, desiredSecrets, nil
}
