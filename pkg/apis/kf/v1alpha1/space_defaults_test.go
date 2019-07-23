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
	"strings"
)

func ExampleSpace_SetDefaults() {
	space := Space{}
	space.Namespace = "mynamespace"
	space.SetDefaults(context.Background())

	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		domainNames = append(domainNames, domain.Domain)
	}

	fmt.Println("Builder:", space.Spec.BuildpackBuild.BuilderImage)
	fmt.Println("Domains:", strings.Join(domainNames, ", "))

	// Output: Builder: gcr.io/kf-releases/buildpack-builder:latest
	// Domains: mynamespace.kf.cluster.local
}

func ExampleSpaceSpecExecution_SetDefaults_dedupe() {
	space := Space{}
	space.Spec.Execution = SpaceSpecExecution{
		Domains: []SpaceDomain{
			{Domain: "example.com"},
			{Domain: "other-example.com"},
			{Domain: "example.com"},
			{Domain: "other-example.com"},
		},
	}
	space.SetDefaults(context.Background())

	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		domainNames = append(domainNames, domain.Domain)
	}

	fmt.Println(strings.Join(domainNames, ", "))

	// Output: example.com, other-example.com
}
