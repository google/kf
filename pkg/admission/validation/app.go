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
	"context"
	"fmt"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfinformer "github.com/google/kf/v2/pkg/client/kf/informers/externalversions/kf/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

// AppValidationCallback is executed to validate App info that requires runtime
// lookups.
func AppValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	app := &v1alpha1.App{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, app); err != nil {
		return err
	}

	if apis.IsInUpdate(ctx) {
		oldApp := &v1alpha1.App{}
		if uns, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apis.GetBaseline(ctx)); err == nil {
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uns, oldApp); err == nil {
				if equality.Semantic.DeepEqual(app.Spec, oldApp.Spec) {
					// Don't validate no-change updates.
					return nil
				}
			}
		}
	}

	spaceInformer := ctx.Value(SpaceInformerKey{}).(kfinformer.SpaceInformer)
	spaceLister := spaceInformer.Lister()

	space, err := spaceLister.Get(app.Namespace)
	if err != nil {
		return err
	}

	// Individual tests go after this comment.

	if err := validateAppDomains(space, app); err != nil {
		return err
	}

	return nil
}

// validateAppDomains validates that the domains on an App's routes are
// valid for the Space.
//
// This validation isn't strictly necessary because a Route will eventually fail
// and become unhealthy if a Space stops including its domain. However, the UX
// is more CF-y to fail fast.
func validateAppDomains(space *v1alpha1.Space, app *v1alpha1.App) error {
	allowedDomains := sets.NewString()
	for _, domain := range space.Status.NetworkConfig.Domains {
		allowedDomains.Insert(domain.Domain)
	}

	appDomains := sets.NewString()
	for _, route := range app.Spec.Routes {
		if route.Domain != "" {
			appDomains.Insert(route.Domain)
		}
	}

	if disallowed := appDomains.Difference(allowedDomains); disallowed.Len() > 0 {
		// List() on a set is sorted.
		validList := strings.Join(allowedDomains.List(), ", ")
		invalidList := strings.Join(disallowed.List(), ", ")
		return fmt.Errorf(
			"Route binding(s) have invalid domain(s): [%s] Space %q only allows domain(s): [%s]",
			invalidList,
			space.Name,
			validList,
		)
	}

	return nil
}
