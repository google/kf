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
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleBuildServiceAccountName() {
	space := &v1alpha1.Space{}
	space.Status.BuildConfig.ServiceAccount = "some-name"
	fmt.Println(BuildServiceAccountName(space))

	// Output: some-name
}

func ExampleBuildImagePushSecretName() {
	secret := &corev1.Secret{}
	secret.ObjectMeta.Name = "gcr-key"
	fmt.Println(BuildImagePushSecretName(secret))

	// Output: kf-registry-gcr-key
}

func ExampleMakeBuildServiceAccount_emptyGSA() {
	sa, secrets, err := MakeBuildServiceAccount(
		&v1alpha1.Space{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-space",
			},
			Status: v1alpha1.SpaceStatus{
				BuildConfig: v1alpha1.SpaceStatusBuildConfig{
					ServiceAccount: "build-creds",
				},
			},
		},
		[]*corev1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gcr-key",
				},
				Data: map[string][]byte{
					"key-1": []byte("value-1"),
					"key-2": []byte("value-2"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ar-key",
				},
				Data: map[string][]byte{
					"key-3": []byte("value-3"),
					"key-4": []byte("value-4"),
				},
			},
		},
		"",
		"ContainerRegistry",
	)

	if err != nil {
		panic(err)
	}

	saSecretNames := []string{}
	for _, s := range sa.Secrets {
		saSecretNames = append(saSecretNames, s.Name)
	}
	saImagePullSecretNames := []string{}
	for _, s := range sa.ImagePullSecrets {
		saImagePullSecretNames = append(saImagePullSecretNames, s.Name)
	}
	fmt.Println("ServiceAccount Name:", sa.Name)
	fmt.Println("ServiceAccount Namespace:", sa.Namespace)
	fmt.Println("ServiceAccount Managed Label:", sa.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("ServiceAccount Annotations:", sa.Annotations)
	fmt.Println("ServiceAccount Secrets:", strings.Join(saSecretNames, ", "))
	fmt.Println("ServiceAccount ImagePullSecrets:", strings.Join(saImagePullSecretNames, ", "))
	fmt.Println("Secrets Count:", len(secrets))
	fmt.Println("Secret Name:", secrets[0].Name)
	fmt.Println("Secret Type:", secrets[0].Type)
	fmt.Println("Secret Namespace:", secrets[0].Namespace)
	fmt.Println("Secret Managed Label:", secrets[0].Labels[v1alpha1.ManagedByLabel])
	fmt.Println("Secret Data[key-1]:", string(secrets[0].Data["key-1"]))
	fmt.Println("Secret Data[key-2]:", string(secrets[0].Data["key-2"]))
	fmt.Println("Secret Annotation:", string(secrets[0].Annotations["tekton.dev/docker-0"]))
	fmt.Println()
	fmt.Println("Secret Name:", secrets[1].Name)
	fmt.Println("Secret Type:", secrets[1].Type)
	fmt.Println("Secret Namespace:", secrets[1].Namespace)
	fmt.Println("Secret Managed Label:", secrets[1].Labels[v1alpha1.ManagedByLabel])
	fmt.Println("Secret Data[key-3]:", string(secrets[1].Data["key-3"]))
	fmt.Println("Secret Data[key-4]:", string(secrets[1].Data["key-4"]))
	fmt.Println("Secret Annotation:", string(secrets[1].Annotations["tekton.dev/docker-1"]))

	// Output: ServiceAccount Name: build-creds
	// ServiceAccount Namespace: some-space
	// ServiceAccount Managed Label: kf
	// ServiceAccount Annotations: map[]
	// ServiceAccount Secrets: kf-registry-gcr-key, kf-registry-ar-key
	// ServiceAccount ImagePullSecrets: kf-registry-gcr-key, kf-registry-ar-key
	// Secrets Count: 2
	// Secret Name: kf-registry-gcr-key
	// Secret Type: kubernetes.io/dockerconfigjson
	// Secret Namespace: some-space
	// Secret Managed Label: kf
	// Secret Data[key-1]: value-1
	// Secret Data[key-2]: value-2
	// Secret Annotation: ContainerRegistry
	//
	// Secret Name: kf-registry-ar-key
	// Secret Type: kubernetes.io/dockerconfigjson
	// Secret Namespace: some-space
	// Secret Managed Label: kf
	// Secret Data[key-3]: value-3
	// Secret Data[key-4]: value-4
	// Secret Annotation: ContainerRegistry
}

func ExampleMakeBuildServiceAccount_withGSA() {
	sa, _, err := MakeBuildServiceAccount(
		&v1alpha1.Space{
			Status: v1alpha1.SpaceStatus{
				BuildConfig: v1alpha1.SpaceStatusBuildConfig{
					ServiceAccount: "build-creds",
				},
			},
		},
		nil,
		"some-gsa",
		"",
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("ServiceAccount Annotations:", sa.Annotations)
	fmt.Println("ServiceAccount len(Secrets):", len(sa.Secrets))

	// Output: ServiceAccount Annotations: map[iam.gke.io/gcp-service-account:some-gsa]
	// ServiceAccount len(Secrets): 0
}
