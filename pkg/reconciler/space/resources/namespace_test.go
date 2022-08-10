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
	"github.com/google/kf/v2/pkg/apis/networking"
	v1 "k8s.io/api/core/v1"
)

func ExampleNamespaceName() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	fmt.Println(NamespaceName(space))

	// Output: my-space
}

func ExampleMakeNamespace() {
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	ns, err := MakeNamespace(space, "some-asm-rev")
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", NamespaceName(space))
	fmt.Println("Label Count:", len(ns.Labels))
	fmt.Println("Managed By:", ns.Labels[managedByLabel])
	fmt.Println("Metadata name:", ns.Labels[v1.LabelMetadataName])
	fmt.Println("Istio Injection:", ns.Labels[networking.IstioInjectionLabel])

	// Output: Name: my-space
	// Label Count: 3
	// Managed By: kf
	// Metadata name: my-space
	// Istio Injection: some-asm-rev
}
