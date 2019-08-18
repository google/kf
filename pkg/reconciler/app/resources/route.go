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
	"hash/crc64"
	"path"
	"regexp"
	"strconv"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"knative.dev/pkg/kmeta"
)

var (
	regexpRouteLabels = regexp.MustCompile(`[a-zA-Z-0-9._-]`)
)

// MakeRouteLabels creates Labels that can be used to tie a Route to a
// VirtualService.
func MakeRouteLabels(spec v1alpha1.RouteSpecFields) map[string]string {
	return map[string]string{
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "route",
		v1alpha1.RouteHostname:  spec.Hostname,
		v1alpha1.RouteDomain:    spec.Domain,
		v1alpha1.RoutePath:      toBase36(path.Join("/", spec.Path)),
	}
}

// MakeRouteAppLabels creates Labels that can be used to lookup the Route for
// the app.
func MakeRouteAppLabels(app *v1alpha1.App) map[string]string {
	return map[string]string{
		v1alpha1.RouteAppName: app.Name,
	}
}

func toBase36(s string) string {
	return strconv.FormatUint(
		crc64.Checksum(
			[]byte(s),
			crc64.MakeTable(crc64.ECMA),
		),
		36)
}

func mustRequirement(key string, op selection.Operator, val string) labels.Requirement {
	r, err := labels.NewRequirement(key, op, []string{val})
	if err != nil {
		panic(err)
	}
	return *r
}

// MakeRouteSelector creates a labels.Selector for listing all the
// corresponding Routes excluding Path.
func MakeRouteSelectorNoPath(spec v1alpha1.RouteSpecFields) labels.Selector {
	return labels.NewSelector().Add(
		mustRequirement(v1alpha1.RouteHostname, selection.Equals, spec.Hostname),
		mustRequirement(v1alpha1.RouteDomain, selection.Equals, spec.Domain),
	)
}

// MakeRouteSelector creates a labels.Selector for listing all the
// corresponding Routes.
func MakeRouteSelector(spec v1alpha1.RouteSpecFields) labels.Selector {
	return labels.NewSelector().Add(
		mustRequirement(v1alpha1.RouteHostname, selection.Equals, spec.Hostname),
		mustRequirement(v1alpha1.RouteDomain, selection.Equals, spec.Domain),
		mustRequirement(v1alpha1.RoutePath, selection.Equals, toBase36(path.Join("/", spec.Path))),
	)
}

// MakeRouteAppSelector creates a labels.Selector for listing all the Routes
// for the given App.
func MakeRouteAppSelector(app *v1alpha1.App) labels.Selector {
	return labels.NewSelector().Add(
		mustRequirement(v1alpha1.RouteAppName, selection.Equals, app.Name),
	)
}

// MakeRoutes creates a Route for the given application.
func MakeRoutes(
	app *v1alpha1.App,
	space *v1alpha1.Space,
) (
	[]v1alpha1.Route,
	[]v1alpha1.RouteClaim,
	error,
) {
	var (
		routes []v1alpha1.Route
		claims []v1alpha1.RouteClaim
	)
	for _, appRoute := range app.Spec.Routes {
		appRoute := appRoute.DeepCopy()
		appRoute.SetSpaceDefaults(space)

		routes = append(routes, v1alpha1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name: v1alpha1.GenerateRouteName(
					appRoute.Hostname,
					appRoute.Domain,
					appRoute.Path,
					app.Name,
				),
				Namespace: space.Name,
				Labels: UnionMaps(
					app.GetLabels(),
					MakeRouteLabels(*appRoute),
					MakeRouteAppLabels(app),
					app.ComponentLabels("route"),
				),
				OwnerReferences: []metav1.OwnerReference{
					*kmeta.NewControllerRef(app),
				},
			},
			Spec: v1alpha1.RouteSpec{
				AppName:         app.Name,
				RouteSpecFields: *appRoute,
			},
		})

		// Claim route
		claims = append(claims, v1alpha1.RouteClaim{
			ObjectMeta: metav1.ObjectMeta{
				Labels: MakeRouteLabels(*appRoute),
				Name: v1alpha1.GenerateRouteClaimName(
					appRoute.Hostname,
					appRoute.Domain,
					appRoute.Path,
				),
				Namespace: space.Name,
			},
			Spec: v1alpha1.RouteClaimSpec{
				RouteSpecFields: *appRoute,
			},
		})
	}

	return routes, claims, nil
}

// UnionMaps is similar to github.com/knative/serving/pkg/resources however it
// takes multiple maps instead of only 2.
func UnionMaps(maps ...map[string]string) map[string]string {
	var result map[string]string

	for _, m := range maps {
		result = resources.UnionMaps(result, m)
	}

	return result
}
