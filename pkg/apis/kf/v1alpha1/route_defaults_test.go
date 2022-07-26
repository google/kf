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

package v1alpha1

import (
	"context"
	"fmt"
)

func ExampleGenerateRouteName() {
	name := GenerateRouteName("some-host", "myns.example.com", "/some/path")
	fmt.Println(name)

	// Output: some-host-myns-example-com--som8ced772f07d7bf61215b525c64ed18bf
}

func ExampleGenerateRouteName_wildcards() {
	name := GenerateRouteName("*", "myns.example.com", "/some/path")
	fmt.Println(name)

	// Output: myns-example-com--some-path4fb7bf8f09018c34a5d240e8398cb069
}

func ExampleRoute_SetDefaults_labels() {
	r := &Route{}
	r.SetDefaults(context.Background())

	fmt.Println("ManagedBy:", r.Labels[ManagedByLabel])
	fmt.Println("Component:", r.Labels[ComponentLabel])

	// Output: ManagedBy: kf
	// Component: route
}
