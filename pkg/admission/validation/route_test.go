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

func TestValidateRouteDomain(t *testing.T) {
	emptySpace := v1alpha1.Space{}
	emptySpace.Name = "empty"

	exampleSpace := v1alpha1.Space{}
	exampleSpace.Name = "example"
	exampleSpace.Status.NetworkConfig.Domains = []v1alpha1.SpaceDomain{
		{Domain: "example.com"},
		{Domain: "some-other-domain.com"},
	}

	exampleRoute := v1alpha1.Route{}
	exampleRoute.Spec.Domain = "example.com"

	badDomainRoute := v1alpha1.Route{}
	badDomainRoute.Spec.Domain = "bad.com"

	cases := map[string]struct {
		space v1alpha1.Space
		route v1alpha1.Route
		want  error
	}{
		"domains match": {
			space: exampleSpace,
			route: exampleRoute,
		},
		"no-domains-on-space": {
			space: emptySpace,
			route: exampleRoute,
			want:  errors.New(`Route has invalid domain: "example.com", Space "empty" only allows domain(s): []`),
		},
		"mismatch": {
			space: exampleSpace,
			route: badDomainRoute,
			want:  errors.New(`Route has invalid domain: "bad.com", Space "example" only allows domain(s): [example.com, some-other-domain.com]`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateRouteDomain(&tc.space, &tc.route)
			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}
