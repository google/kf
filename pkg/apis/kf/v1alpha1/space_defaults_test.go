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

	routecfg "github.com/google/kf/third_party/knative-serving/pkg/reconciler/route/config"
)

func dummyConfig() context.Context {
	cfg := &routecfg.Config{
		Domain: &routecfg.Domain{
			Domains: map[string]*routecfg.LabelSelector{
				"custom.example.com": {},
			},
		},
	}

	return routecfg.ToContext(context.TODO(), cfg)
}

func ExampleSpace_SetDefaults() {
	space := Space{}
	space.Name = "mynamespace"
	space.SetDefaults(dummyConfig())

	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		if domain.Default {
			domainNames = append(domainNames, "*"+domain.Domain)
			continue
		}
		domainNames = append(domainNames, domain.Domain)
	}

	if _, err := fmt.Println("Builder:", space.Spec.BuildpackBuild.BuilderImage); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("Domains:", strings.Join(domainNames, ", ")); err != nil {
		panic(err)
	}

	// Output: Builder: gcr.io/kf-releases/buildpack-builder:latest
	// Domains: *mynamespace.custom.example.com
}

func ExampleSpaceSpecExecution_SetDefaults_dedupe() {
	space := Space{}
	space.Spec.Execution = SpaceSpecExecution{
		Domains: []SpaceDomain{
			{Domain: "example.com"},
			{Domain: "other-example.com"},
			{Domain: "example.com", Default: true},
			{Domain: "other-example.com"},
		},
	}
	space.SetDefaults(context.Background())

	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		if domain.Default {
			domainNames = append(domainNames, "*"+domain.Domain)
			continue
		}
		domainNames = append(domainNames, domain.Domain)
	}

	if _, err := fmt.Println(strings.Join(domainNames, ", ")); err != nil {
		panic(err)
	}

	// Output: *example.com, other-example.com

}

func ExampleSpaceSpecSecurity_SetDefaults() {
	space := Space{}
	space.Spec.Security = SpaceSpecSecurity{
		EnableDeveloperLogsAccess: false,
	}
	space.SetDefaults(dummyConfig())

	if _, err := fmt.Println("EnableDeveloperLogsAccess:", space.Spec.Security.EnableDeveloperLogsAccess); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("BuildServiceAccount:", space.Spec.Security.BuildServiceAccount); err != nil {
		panic(err)
	}

	// Output: EnableDeveloperLogsAccess: true
	// BuildServiceAccount: kf-builder
}

func ExampleSpaceSpecSecurity_SetDefaults_preserves() {
	space := Space{}
	space.Spec.Security = SpaceSpecSecurity{
		EnableDeveloperLogsAccess: false,
		BuildServiceAccount:       "some-other-account",
	}
	space.SetDefaults(dummyConfig())

	if _, err := fmt.Println("EnableDeveloperLogsAccess:", space.Spec.Security.EnableDeveloperLogsAccess); err != nil {
		panic(err)
	}
	if _, err := fmt.Println("BuildServiceAccount:", space.Spec.Security.BuildServiceAccount); err != nil {
		panic(err)
	}

	// Output: EnableDeveloperLogsAccess: true
	// BuildServiceAccount: some-other-account
}

func ExampleSpaceSpecExecution_SetDefaults_badContextPanic() {
	space := Space{}
	space.Name = "mynamespace"
	space.SetDefaults(context.Background())

	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		if domain.Default {
			domainNames = append(domainNames, "*"+domain.Domain)
			continue
		}
		domainNames = append(domainNames, domain.Domain)
	}

	if _, err := fmt.Println("Domains:", strings.Join(domainNames, ", ")); err != nil {
		panic(err)
	}

	// Output: Domains: *mynamespace.example.com
}

func ExampleDefaultDomain() {
	if _, err := fmt.Println(DefaultDomain(context.Background())); err != nil {
		panic(err)
	}

	// Output: example.com
}
