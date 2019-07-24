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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeRouteLabels creates labels that can be used to tie a route to a
// VirtualService.
func MakeRouteLabels() map[string]string {
	return map[string]string{
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "route",
	}
}

// MakeRoutes creates a Route for the given application.
func MakeRoutes(app *v1alpha1.App, space *v1alpha1.Space) ([]v1alpha1.Route, error) {
	var routes []v1alpha1.Route
	for _, appRoute := range app.Spec.Routes {
		appRoute := appRoute.DeepCopy()
		appRoute.SetSpaceDefaults(space)

		routes = append(routes, v1alpha1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v1alpha1.GenerateRouteName(appRoute.Hostname, appRoute.Domain, appRoute.Path),
				Namespace: app.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*kmeta.NewControllerRef(space),
				},
				Labels: resources.UnionMaps(app.GetLabels(), MakeRouteLabels()),
			},
			Spec: v1alpha1.RouteSpec{
				AppNames:        []string{app.Name},
				RouteSpecFields: *appRoute,
			},
		})
	}

	return routes, nil
}
