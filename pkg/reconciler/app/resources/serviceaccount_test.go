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
	"fmt"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleServiceAccountName() {
	app := &v1alpha1.App{}
	app.Name = "some-app"
	fmt.Println(ServiceAccountName(app))

	// Output: sa-some-app
}

func ExampleMakeServiceAccount_wiEnabled() {
	sa := MakeServiceAccount(
		&v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-app",
				Namespace: "some-space",
			},
		},
		[]corev1.LocalObjectReference{},
	)

	secretNames := []string{}
	for _, s := range sa.ImagePullSecrets {
		secretNames = append(secretNames, s.Name)
	}

	fmt.Println("ServiceAccount Name:", sa.Name)
	fmt.Println("ServiceAccount Namespace:", sa.Namespace)
	fmt.Println("ServiceAccount Managed Label:", sa.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("ServiceAccount Component Label:", sa.Labels[v1alpha1.ComponentLabel])
	fmt.Println("ServiceAccount ImagePullSecrets:", strings.Join(secretNames, ", "))

	// Output: ServiceAccount Name: sa-some-app
	// ServiceAccount Namespace: some-space
	// ServiceAccount Managed Label: kf
	// ServiceAccount Component Label: serviceaccount
	// ServiceAccount ImagePullSecrets:
}

func ExampleMakeServiceAccount_wiDisabled() {
	sa := MakeServiceAccount(
		&v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-app",
				Namespace: "some-space",
			},
		},
		[]corev1.LocalObjectReference{
			{Name: "kf-builder-ar"},
			{Name: "kf-builder-gcr"},
		},
	)

	secretNames := []string{}
	for _, s := range sa.ImagePullSecrets {
		secretNames = append(secretNames, s.Name)
	}

	fmt.Println("ServiceAccount Name:", sa.Name)
	fmt.Println("ServiceAccount Namespace:", sa.Namespace)
	fmt.Println("ServiceAccount Managed Label:", sa.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("ServiceAccount Component Label:", sa.Labels[v1alpha1.ComponentLabel])
	fmt.Println("ServiceAccount ImagePullSecrets:", strings.Join(secretNames, ", "))

	// Output: ServiceAccount Name: sa-some-app
	// ServiceAccount Namespace: some-space
	// ServiceAccount Managed Label: kf
	// ServiceAccount Component Label: serviceaccount
	// ServiceAccount ImagePullSecrets: kf-builder-ar, kf-builder-gcr
}

func ExampleFilterAndSortKfSecrets() {
	nonKfSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-token",
		},
	}

	kfBuildSecretGcr := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kf-registry-gcr-key",
			Labels: map[string]string{
				v1alpha1.ManagedByLabel: "kf",
			},
		},
	}

	kfBuildSecretAr := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kf-registry-ar-key",
			Labels: map[string]string{
				v1alpha1.ManagedByLabel: "kf",
			},
		},
	}

	secrets := []*corev1.Secret{nonKfSecret, kfBuildSecretGcr, kfBuildSecretAr}
	filteredSecrets := FilterAndSortKfSecrets(secrets)

	fmt.Println("All secrets:", strings.Join(secretNames(secrets), ", "))
	fmt.Println("Filtered and sorted Kf secrets:", strings.Join(secretNames(filteredSecrets), ", "))

	// Output: All secrets: default-token, kf-registry-gcr-key, kf-registry-ar-key
	// Filtered and sorted Kf secrets: kf-registry-ar-key, kf-registry-gcr-key
}

func secretNames(secrets []*corev1.Secret) []string {
	names := []string{}
	for _, s := range secrets {
		names = append(names, s.Name)
	}
	return names
}
