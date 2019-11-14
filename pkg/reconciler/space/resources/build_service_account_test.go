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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleBuildServiceAccountName() {
	space := &v1alpha1.Space{}
	space.Spec.Security.BuildServiceAccount = "some-name"
	if _, err := fmt.Println(BuildServiceAccountName(space)); err != nil {
		panic(err)
	}

	// Output: kf-builder
}

func ExampleBuildSecretName() {
	space := &v1alpha1.Space{}
	space.Spec.Security.BuildServiceAccount = "some-name"
	if _, err := fmt.Println(BuildSecretName(space)); err != nil {
		panic(err)
	}

	// Output: kf-builder
}

func ExampleMakeBuildServiceAccount() {
	sa, secrets, err := MakeBuildServiceAccount(
		&v1alpha1.Space{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-space",
			},
			Spec: v1alpha1.SpaceSpec{
				Security: v1alpha1.SpaceSpecSecurity{
					BuildServiceAccount: "build-creds",
				},
			},
		},
		[]*corev1.Secret{
			{
				Data: map[string][]byte{
					"key-1": []byte("value-1"),
					"key-2": []byte("value-2"),
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	if _, err := fmt.Println("ServiceAccount Name:", sa.Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("ServiceAccount Namespace:", sa.Namespace); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("ServiceAccount Managed Label:", sa.Labels[v1alpha1.ManagedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("ServiceAccount Secret:", sa.Secrets[0].Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secrets Count:", len(secrets)); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Name:", secrets[0].Name); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Type:", secrets[0].Type); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Namespace:", secrets[0].Namespace); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Managed Label:", secrets[0].Labels[v1alpha1.ManagedByLabel]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Data[key-1]:", string(secrets[0].Data["key-1"])); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Secret Data[key-2]:", string(secrets[0].Data["key-2"])); err != nil {
		panic(err)
	}

	// Output: ServiceAccount Name: kf-builder
	// ServiceAccount Namespace: some-space
	// ServiceAccount Managed Label: kf
	// ServiceAccount Secret: kf-builder
	// Secrets Count: 1
	// Secret Name: kf-builder
	// Secret Type: kubernetes.io/dockerconfigjson
	// Secret Namespace: some-space
	// Secret Managed Label: kf
	// Secret Data[key-1]: value-1
	// Secret Data[key-2]: value-2
}
