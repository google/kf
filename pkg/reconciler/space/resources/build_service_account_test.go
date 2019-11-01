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
	fmt.Println(BuildServiceAccountName(space))

	// Output: some-name
}

func ExampleBuildSecretName() {
	space := &v1alpha1.Space{}
	space.Spec.Security.BuildServiceAccount = "some-name"
	fmt.Println(BuildSecretName(space))

	// Output: some-name
}

func ExampleMakeBuildServiceAccount() {
	sa, secret, err := MakeBuildServiceAccount(
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
		&corev1.Secret{
			Data: map[string][]byte{
				"key-1": []byte("value-1"),
				"key-2": []byte("value-2"),
			},
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("ServiceAccount Name:", sa.Name)
	fmt.Println("ServiceAccount Namespace:", sa.Namespace)
	fmt.Println("ServiceAccount Managed Label:", sa.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("ServiceAccount Secret:", sa.Secrets[0].Name)
	fmt.Println("Secret Name:", secret.Name)
	fmt.Println("Secret Type:", secret.Type)
	fmt.Println("Secret Namespace:", secret.Namespace)
	fmt.Println("Secret Managed Label:", secret.Labels[v1alpha1.ManagedByLabel])
	fmt.Println("Secret Data[key-1]:", string(secret.Data["key-1"]))
	fmt.Println("Secret Data[key-2]:", string(secret.Data["key-2"]))

	// Output: ServiceAccount Name: build-creds
	// ServiceAccount Namespace: some-space
	// ServiceAccount Managed Label: kf
	// ServiceAccount Secret: build-creds
	// Secret Name: build-creds
	// Secret Type: kubernetes.io/dockerconfigjson
	// Secret Namespace: some-space
	// Secret Managed Label: kf
	// Secret Data[key-1]: value-1
	// Secret Data[key-2]: value-2
}
