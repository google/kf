// Copyright 2020 Google LLC
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
	"math"

	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmp"
)

// Validate implements apis.Validatable.
func (binding *ServiceInstanceBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(binding.GetObjectMeta()).ViaField("metadata"))

	// Deny changes to spec except UnbindRequests
	if apis.IsInUpdate(ctx) {
		original := apis.GetBaseline(ctx).(*ServiceInstanceBinding)
		binding.Spec.UnbindRequests = original.Spec.UnbindRequests
		if diff, err := kmp.ShortDiff(original.Spec, binding.Spec); err != nil {
			return errs.Also(&apis.FieldError{
				Message: "Failed to diff",
				Paths:   []string{"spec"},
				Details: err.Error(),
			})
		} else if diff != "" {
			return errs.Also(&apis.FieldError{
				Message: "Immutable fields changed (-old +new)",
				Paths:   []string{"spec"},
				Details: diff,
			})
		}
	} else {
		errs = errs.Also(binding.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	}

	return errs
}

// Validate implements apis.Validatable.
func (spec *ServiceInstanceBindingSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(spec.BindingType.Validate(ctx))

	if spec.InstanceRef.Name == "" {
		errs = errs.Also(apis.ErrMissingField("instanceRef.name"))
	}

	if spec.ParametersFrom.Name == "" {
		errs = errs.Also(apis.ErrMissingField("parametersFrom.name"))
	}

	// Test < 0 rather than <= 0 to ensure configurations prior to the field
	// addition are valid. The default function will upgrade those in the
	// reconciler.
	if v := spec.ProgressDeadlineSeconds; v < 0 {
		errs = errs.Also(apis.ErrOutOfBoundsValue(v, 1, math.MaxInt64, "progressDeadlineSeconds"))
	}

	return
}

// Validate implements apis.Validatable.
func (bindingType *BindingType) Validate(ctx context.Context) (errs *apis.FieldError) {
	defined := sets.NewString()
	all := sets.NewString()

	fields := []struct {
		fieldName string
		isNil     bool
		validator apis.Validatable
	}{
		{
			fieldName: "app",
			isNil:     bindingType.App == nil,
			validator: bindingType.App,
		},
		{
			fieldName: "route",
			isNil:     bindingType.Route == nil,
			validator: bindingType.Route,
		},
	}

	for _, field := range fields {
		all.Insert(field.fieldName)

		if !field.isNil {
			defined.Insert(field.fieldName)
			errs = errs.Also(field.validator.Validate(ctx).ViaField(field.fieldName))
		}
	}

	switch defined.Len() {
	case 0: // missing one
		return apis.ErrMissingOneOf(all.List()...)
	case 1: // return the exact errors
		return errs
	default: // multiple defined
		return apis.ErrMultipleOneOf(defined.List()...)
	}
}

// Validate implements apis.Validatable.
func (app *AppRef) Validate(ctx context.Context) (errs *apis.FieldError) {
	if app.Name == "" {
		errs = errs.Also(apis.ErrMissingField("appName"))
	}

	return
}

// Validate implements apis.Validatable.
func (route *RouteRef) Validate(ctx context.Context) (errs *apis.FieldError) {
	if route.Domain == "" {
		errs = errs.Also(apis.ErrMissingField("routeDomain"))
	}

	return
}
