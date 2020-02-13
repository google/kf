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

package completion

import (
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ExamplePrintNames() {
	a := unstructured.Unstructured{}
	a.SetName("app-a")

	z := unstructured.Unstructured{}
	z.SetName("app-z")

	PrintNames(os.Stdout, &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{z, a},
	})

	// Output: app-a app-z
}
