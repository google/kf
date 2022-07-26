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

	"github.com/google/kf/v2/pkg/apis/kf"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/validation"
	"knative.dev/pkg/apis"
)

const (
	KfNamespace = "kf"
)

type skipDestinationPortCheck struct{}

var skipDestinationPortCheckKey = skipDomainCheck{}

func withAllowEmptyDestinationPort(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipDestinationPortCheckKey, true)
}

func allowEmptyDestinationPort(ctx context.Context) bool {
	return ctx.Value(skipDestinationPortCheckKey) != nil
}

// Validate makes sure that RouteWeightBinding is properly configured.
func (r *RouteWeightBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	if r.Weight == nil {
		errs = errs.Also(apis.ErrMissingField("weight"))
	} else if *r.Weight < 0 {
		errs = errs.Also(apis.ErrInvalidValue(*r.Weight, "weight"))
	}

	if !allowEmptyDestinationPort(ctx) && r.DestinationPort == nil {
		errs = errs.Also(apis.ErrMissingField("destinationPort"))
	}

	if r.DestinationPort != nil {
		errs = errs.Also(kf.ValidatePortNumberBounds(*r.DestinationPort, "destinationPort"))
	}

	// don't include a ViaField because the field is embedded
	return errs.Also(r.RouteSpecFields.Validate(ctx))
}

type skipDomainCheck struct{}

var skipDomainCheckKey = skipDomainCheck{}

func withAllowEmptyDomains(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipDomainCheckKey, true)
}

func allowEmptyDomains(ctx context.Context) bool {
	return ctx.Value(skipDomainCheckKey) != nil
}

// Validate makes sure that RouteSpecFields is properly configured.
func (r *RouteSpecFields) Validate(ctx context.Context) (errs *apis.FieldError) {

	if !allowEmptyDomains(ctx) && r.Domain == "" {
		errs = errs.Also(apis.ErrMissingField("domain"))
	}

	switch r.Hostname {
	case "":
		break // Hostname is optional

	case "*":
		break // Explicitly allowed to indicate wildcard routes

	default:
		for _, errMsg := range validation.IsDNS1123Label(r.Hostname) {
			errs = errs.Also(&apis.FieldError{
				Message: "Invalid Value",
				Details: errMsg,
				Paths:   []string{"hostname"},
			})
		}
	}

	if _, err := BuildPathRegexp(r.Path); err != nil {
		errs = errs.Also(apis.ErrInvalidValue(r.Path, "path"))
	}

	return errs
}

// Validate validates a Route.
func (r *Route) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(r.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(r.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate validates a RouteSpec.
func (r *RouteSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	// don't include a ViaField because the field is embedded
	return errs.Also(r.RouteSpecFields.Validate(ctx))
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
