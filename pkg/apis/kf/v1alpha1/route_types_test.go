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

import "fmt"

func ExampleRouteSpecFields_String() {
	r := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
		Path:     "bar",
	}

	if _, err := fmt.Println(r.String()); err != nil {
		panic(err)
	}

	// Output: foo.example.com/bar
}

func ExampleRouteSpecFields_String_without_hostname() {
	r := RouteSpecFields{
		Domain: "example.com",
		Path:   "bar",
	}

	if _, err := fmt.Println(r.String()); err != nil {
		panic(err)
	}

	// Output: example.com/bar
}

func ExampleRouteSpecFields_String_without_path() {
	r := RouteSpecFields{
		Hostname: "foo",
		Domain:   "example.com",
	}

	if _, err := fmt.Println(r.String()); err != nil {
		panic(err)
	}

	// Output: foo.example.com/
}
