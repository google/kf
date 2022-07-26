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
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/ptr"
)

// MakeRouteLabels creates Labels that can be used to tie a Route to a
// VirtualService.
func MakeRouteLabels() map[string]string {
	return map[string]string{
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "route",
	}
}

// MakeRoutes creates a Route for the given application.
func MakeRoutes(
	app *v1alpha1.App,
	space *v1alpha1.Space,
) (
	claims []v1alpha1.Route,
	bindings []v1alpha1.QualifiedRouteBinding,
	err error,
) {
	for _, binding := range app.Spec.Routes {
		binding := binding.DeepCopy()
		if app.Spec.Instances.Stopped == true {
			binding.Weight = ptr.Int32(0)
		}

		bindings = append(bindings, binding.Qualify(space.DefaultDomainOrBlank(), app.Name))
	}

	// Merge bindings so after the default domain has been applied any that are
	// now equal get merged.
	bindings = v1alpha1.MergeQualifiedBindings(bindings)

	// Build claims, only one claim per name will be built
	claimNames := sets.NewString()
	for _, binding := range bindings {
		name := v1alpha1.GenerateRouteName(
			binding.Source.Hostname,
			binding.Source.Domain,
			binding.Source.Path,
		)

		if claimNames.Has(name) {
			continue
		}
		claimNames.Insert(name)

		claims = append(claims, v1alpha1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    MakeRouteLabels(),
				Name:      name,
				Namespace: space.Name,
			},
			Spec: v1alpha1.RouteSpec{
				RouteSpecFields: binding.Source,
			},
		})
	}

	return claims, bindings, nil
}
