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

	"github.com/gorilla/mux"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	KfNamespace = "kf"
)

// Validate makes sure that Route is properly configured.
func (r *Route) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	if r.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}

	errs = errs.Also(r.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	// If we have errors, bail. No need to do the network call.
	if errs.Error() != "" {
		return errs
	}

	return checkVirtualServiceCollision(ctx, r.Spec.Hostname, r.Spec.Domain, r.GetNamespace(), errs)
}

func checkVirtualServiceCollision(ctx context.Context, hostname, domain, namespace string, errs *apis.FieldError) *apis.FieldError {
	// XXX: We probably shouldn't be fetching VirtualServices in a webhook,
	// however we need to ensure the resulting VirtualService doesn't
	// conflict.
	vs, err := IstioClientFromContext(ctx).
		VirtualServices(KfNamespace).
		Get(GenerateName(hostname, domain), metav1.GetOptions{})

	if apierrs.IsNotFound(err) {
		vs = nil
	} else if err != nil {
		return errs.Also(&apis.FieldError{
			Message: "failed to validate hostname + domain collisions",
			Details: fmt.Sprintf("failed to fetch VirtualServices: %s", err),
		})
	}

	if vs != nil && vs.Annotations["space"] != namespace {
		errs = errs.Also(&apis.FieldError{
			Message: "Immutable field changed",
			Paths:   []string{"namespace"},
			Details: fmt.Sprintf("The route is invalid: Routes for this host and domain have been reserved for another space."),
		})
	}

	return errs
}

// Validate makes sure that RouteSpec is properly configured.
func (r *RouteSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	if r.AppName == "" {
		errs = errs.Also(apis.ErrMissingField("appName"))
	}

	return errs.Also(r.RouteSpecFields.Validate(ctx).ViaField("routeSpecFields"))
}

// Validate makes sure that RouteSpecFields is properly configured.
func (r *RouteSpecFields) Validate(ctx context.Context) (errs *apis.FieldError) {

	if r.Domain == "" {
		errs = errs.Also(apis.ErrMissingField("domain"))
	}

	if r.Hostname == "www" {
		errs = errs.Also(apis.ErrInvalidValue("hostname", r.Hostname))
	}

	if _, err := BuildPathRegexp(r.Path); err != nil {
		errs = errs.Also(apis.ErrInvalidValue("path", r.Path))
	}

	return errs
}

// Validate validates a RouteClaim.
func (r *RouteClaim) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	if r.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}

	errs = errs.Also(r.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	// If we have errors, bail. No need to do the network call.
	if errs.Error() != "" {
		return errs
	}

	return checkVirtualServiceCollision(ctx, r.Spec.Hostname, r.Spec.Domain, r.GetNamespace(), errs)
}

// Validate validates a RouteClaimSpec.
func (r *RouteClaimSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	return errs.Also(r.RouteSpecFields.Validate(ctx).ViaField("routeSpecFields"))
}

// BuildPathRegexp uses gorilla/mux to convert a path into regular expression
// that can be used to determine if a requests' path matches.
func BuildPathRegexp(path string) (string, error) {
	// If its just the root path, we'll add that back as optional
	if path == "/" {
		path = ""
	}

	p, err := (&mux.Router{}).PathPrefix(path).GetPathRegexp()
	if err != nil {
		return "", err
	}
	return p + `(/.*)?`, nil
}
