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

func ExampleRoute_SetDefaults_prefixRoutes() {
	r := &Route{}
	r.Spec.Path = "some-path"
	r.SetDefaults(context.Background())

	if _, err := fmt.Println("Route:", r.Spec.Path); err != nil {
		panic(err)
	}

	// Output: Route: /some-path
}

func ExampleRoute_SetDefaults_labels() {
	r := &Route{}
	r.Spec.Hostname = "some-hostname"
	r.Spec.Domain = "example.com"
	r.SetDefaults(context.Background())

	if _, err := fmt.Println("Hostname:", r.Labels[RouteHostname]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Domain:", r.Labels[RouteDomain]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Path:", r.Labels[RoutePath]); err != nil {
		panic(err)
	}

	// Output: Hostname: some-hostname
	// Domain: example.com
	// Path: pvdf1ls1w14a
}

func ExampleRouteClaim_SetDefaults_labels() {
	r := &RouteClaim{}
	r.Spec.Hostname = "some-hostname"
	r.Spec.Domain = "example.com"
	r.SetDefaults(context.Background())

	if _, err := fmt.Println("Hostname:", r.Labels[RouteHostname]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Domain:", r.Labels[RouteDomain]); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Path:", r.Labels[RoutePath]); err != nil {
		panic(err)
	}

	// Output: Hostname: some-hostname
	// Domain: example.com
	// Path: pvdf1ls1w14a
}
