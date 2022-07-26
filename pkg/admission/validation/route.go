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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

// RouteValidationCallback is executed to validate Route info that requires
// runtime lookups.
func RouteValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	// Sanity check to see that we're in create; blocking due to updates is
	// dangerous because knative/pkg doesn't differentiate between spec and
	// status updates for callbacks meaning resources can get stuck.
	if !apis.IsInCreate(ctx) {
		return nil
	}

	route := &v1alpha1.Route{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, route); err != nil {
		return err
	}

	spaceInformer := ctx.Value(SpaceInformerKey{}).(kfinformer.SpaceInformer)
	spaceLister := spaceInformer.Lister()

	space, err := spaceLister.Get(route.Namespace)
	if err != nil {
		return err
	}

	// Individual tests go after this comment.

	if err := validateRouteDomain(space, route); err != nil {
		return err
	}

	return nil
}

// validateRouteDomain validates that the domains on a Route are
// valid for the Space.
//
// This validation isn't strictly necessary because a Route will eventually fail
// and become unhealthy if a Space stops including its domain. However, the UX
// is more CF-y to fail fast.
func validateRouteDomain(space *v1alpha1.Space, route *v1alpha1.Route) error {
	allowedDomains := sets.NewString()
	for _, domain := range space.Status.NetworkConfig.Domains {
		allowedDomains.Insert(domain.Domain)
	}

	if !allowedDomains.Has(route.Spec.Domain) {
		// List() on a set is sorted.
		validList := strings.Join(allowedDomains.List(), ", ")
		return fmt.Errorf(
			"Route has invalid domain: %q, Space %q only allows domain(s): [%s]",
			route.Spec.Domain,
			space.Name,
			validList,
		)
	}

	return nil
}
