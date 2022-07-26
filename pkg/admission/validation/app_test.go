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

package kfvalidation

import (
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestValidateAppDomains(t *testing.T) {
	emptySpace := v1alpha1.Space{}
	emptySpace.Name = "empty"

	exampleSpace := v1alpha1.Space{}
	exampleSpace.Name = "example"
	exampleSpace.Status.NetworkConfig.Domains = []v1alpha1.SpaceDomain{
		{Domain: "example.com"},
		{Domain: "some-other-domain.com"},
	}

	exampleApp := v1alpha1.App{}
	exampleApp.Spec.Routes = []v1alpha1.RouteWeightBinding{
		{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: "example.com"}},
		{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: ""}},
	}

	badDomainApp := v1alpha1.App{}
	badDomainApp.Spec.Routes = []v1alpha1.RouteWeightBinding{
		{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: "bad2.com"}},
		{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: "bad.com"}},
	}

	cases := map[string]struct {
		space v1alpha1.Space
		app   v1alpha1.App
		want  error
	}{
		"no-domains-no-routes": {
			space: emptySpace,
		},
		"domains-no-routes": {
			space: exampleSpace,
		},
		"domains-routes": {
			space: exampleSpace,
			app:   exampleApp,
		},
		"no-domains-some-routes": {
			space: emptySpace,
			app:   exampleApp,
			want:  errors.New(`Route binding(s) have invalid domain(s): [example.com] Space "empty" only allows domain(s): []`),
		},
		"mismatch": {
			space: exampleSpace,
			app:   badDomainApp,
			want:  errors.New(`Route binding(s) have invalid domain(s): [bad.com, bad2.com] Space "example" only allows domain(s): [example.com, some-other-domain.com]`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateAppDomains(&tc.space, &tc.app)
			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}
