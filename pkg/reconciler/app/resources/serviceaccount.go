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

package resources

import (
	"sort"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// ServiceAccountName generates a name for the app service account.
func ServiceAccountName(app *kfv1alpha1.App) string {
	return v1alpha1.GenerateName("sa", app.Name)
}

// MakeServiceAccount constructs a K8s service account, which is used by the deployment for the app.
func MakeServiceAccount(app *kfv1alpha1.App, imagePullSecrets []corev1.LocalObjectReference) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("serviceaccount")),
		},
		ImagePullSecrets: imagePullSecrets,
	}
}

// FilterAndSortKfSecrets filters a list of secrets and returns the ones that are managed by Kf, sorted alphabetically by name.
func FilterAndSortKfSecrets(secrets []*corev1.Secret) []*corev1.Secret {
	filteredSecrets := []*corev1.Secret{}
	for _, s := range secrets {
		if s.ObjectMeta.Labels[v1alpha1.ManagedByLabel] == "kf" {
			filteredSecrets = append(filteredSecrets, s)
		}
	}
	sort.SliceStable(filteredSecrets, func(i, j int) bool {
		return filteredSecrets[i].Name < filteredSecrets[j].Name
	})
	return filteredSecrets
}
