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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleBuildServiceAccountName() {
	fmt.Println(BuildServiceAccountName(&v1alpha1.Space{}))

	// Output: kf-build-service-account
}

func ExampleMakeBuildServiceAccount() {
	sa, err := MakeBuildServiceAccount(&v1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-space",
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", sa.Name)
	fmt.Println("Namespace:", sa.Namespace)
	fmt.Println("Managed Label:", sa.Labels[v1alpha1.ManagedByLabel])

	// Output: Name: kf-build-service-account
	// Namespace: some-space
	// Managed Label: kf
}
