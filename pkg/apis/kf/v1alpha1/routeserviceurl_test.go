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

package v1alpha1

import (
	"encoding/json"
	"fmt"
)

func ExampleRouteServiceURL_String() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:80",
		Path:   "foo",
	}
	fmt.Println(r.String())

	// Output: http://auth.example.com:80/foo
}

func ExampleRouteServiceURL_String_without_scheme() {
	r := &RouteServiceURL{
		Host: "auth.example.com:80",
		Path: "foo",
	}
	fmt.Println(r.String())

	// Output: //auth.example.com:80/foo
}

func ExampleRouteServiceURL_String_without_port() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com",
		Path:   "foo",
	}
	fmt.Println(r.String())

	// Output: http://auth.example.com/foo
}

func ExampleRouteServiceURL_String_without_path() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:80",
	}
	fmt.Println(r.String())

	// Output: http://auth.example.com:80
}

func ExampleRouteServiceURL_String_path_format() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com",
		Path:   "/foo",
	}
	fmt.Println(r.String())

	// Output: http://auth.example.com/foo
}

func ExampleRouteServiceURL_Port() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:8080",
		Path:   "foo",
	}
	fmt.Println(r.Port())

	// Output: 8080
}

func ExampleRouteServiceURL_Port_empty() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com",
		Path:   "foo",
	}
	fmt.Println(r.Port())

	// Output:
}

func ExampleRouteServiceURL_Hostname() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com",
		Path:   "foo",
	}
	fmt.Println(r.Hostname())

	// Output: auth.example.com
}

func ExampleRouteServiceURL_Hostname_strips_port() {
	r := &RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:8080",
		Path:   "foo",
	}
	fmt.Println(r.Hostname())

	// Output: auth.example.com
}

func ExampleRouteServiceURL_MarshalJSON() {
	r := RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:8080",
		Path:   "foo",
	}
	marshaled, _ := json.Marshal(r)
	fmt.Println(string(marshaled))

	// Output: "http://auth.example.com:8080/foo"
}

func ExampleRouteServiceURL_UnmarshalJSON() {
	r := RouteServiceURL{
		Scheme: "http",
		Host:   "auth.example.com:8080",
		Path:   "foo",
	}
	marshaled, _ := json.Marshal(r)
	rsURL := &RouteServiceURL{}
	json.Unmarshal(marshaled, rsURL)
	fmt.Println(rsURL)
	fmt.Println("Scheme:", rsURL.Scheme)
	fmt.Println("Host:", rsURL.Host)
	fmt.Println("Path:", rsURL.Path)

	// Output: http://auth.example.com:8080/foo
	// Scheme: http
	// Host: auth.example.com:8080
	// Path: /foo
}
