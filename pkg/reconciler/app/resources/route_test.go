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
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeRoutes(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		app    v1alpha1.App
		space  v1alpha1.Space
		assert func(t *testing.T, routes []v1alpha1.Route)
	}{
		"no domain, uses space default": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com", Default: true},
							{Domain: "wrong.example.com", Default: false},
						},
					},
				},
			},
			app: v1alpha1.App{
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: ""},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
			},
		},
		"adds app name in AppNames": {
			app: v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-name",
				},
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: "example.com"},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "route.Spec.Domain", "example.com", routes[0].Spec.Domain)
				testutil.AssertEqual(t, "route.Spec.Hostname", "some-hostname", routes[0].Spec.Hostname)
				testutil.AssertEqual(t, "route.Spec.AppNames", []string{"some-app-name"}, routes[0].Spec.AppNames)
			},
		},
		"ObjectMeta": {
			space: v1alpha1.Space{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-space-name",
				},
			},
			app: v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-app-name",
					Labels:    map[string]string{"a": "1", "b": "2"},
				},
				Spec: v1alpha1.AppSpec{
					Routes: []v1alpha1.RouteSpecFields{
						{Hostname: "some-hostname", Domain: "some-domain", Path: "some-path"},
					},
				},
			},
			assert: func(t *testing.T, routes []v1alpha1.Route) {
				testutil.AssertEqual(t, "len(routes)", 1, len(routes))
				testutil.AssertEqual(t, "route.ObjectMeta.Namespace", "some-namespace", routes[0].ObjectMeta.Namespace)
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.Name",
					v1alpha1.GenerateRouteName("some-hostname", "some-domain", "some-path"),
					routes[0].ObjectMeta.Name,
				)
				testutil.AssertEqual(t, "route.ObjectMeta.Labels", map[string]string{
					"a":                     "1",
					"b":                     "2",
					v1alpha1.ManagedByLabel: "kf",
					v1alpha1.ComponentLabel: "route",
				}, routes[0].ObjectMeta.Labels)

				b := true
				testutil.AssertEqual(
					t,
					"route.ObjectMeta.OwnerReferences",
					[]metav1.OwnerReference{{
						APIVersion:         "kf.dev/v1alpha1",
						Kind:               "Space",
						Name:               "some-space-name",
						Controller:         &b,
						BlockOwnerDeletion: &b,
					}},
					routes[0].ObjectMeta.OwnerReferences,
				)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			routes, err := MakeRoutes(&tc.app, &tc.space)
			testutil.AssertNil(t, "err", err)
			tc.assert(t, routes)
		})
	}
}
