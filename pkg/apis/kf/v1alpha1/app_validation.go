// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/knative/serving/pkg/apis/serving"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

// Validate checks for errors in the App's spec or status fields.
func (app *App) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if !apis.IsInStatusUpdate(ctx) {
		errs = errs.Also(app.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	}

	return errs
}

// Validate checks that the pod template the user has submitted is valid
// and that the scaling and lifecycle is valid.
func (spec *AppSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	errs = errs.Also(ValidatePodSpec(spec.Template.Spec).ViaField("template.spec"))
	errs = errs.Also(spec.Instances.Validate(ctx).ViaField("instances"))
	errs = errs.Also(spec.ValidateSourceSpec(ctx).ViaField("source"))
	errs = errs.Also(spec.ValidateServiceBindings(ctx).ViaField("serviceBindings"))

	return errs
}

// ValidateSourceSpec validates the SourceSpec embedded in the AppSpec.
func (spec *AppSpec) ValidateSourceSpec(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(apis.CheckDisallowedFields(spec.Source, AppSpecSourceMask(spec.Source)))

	// Fail if the app source has changed without changing the UpdateRequests.
	if base := apis.GetBaseline(ctx); base != nil {
		if old, ok := base.(*App); ok {
			previousValue := old.Spec.Source.UpdateRequests
			newValue := spec.Source.UpdateRequests
			if previousValue > newValue {
				msg := fmt.Sprintf("UpdateRequests must be nondecreasing, previous value: %d new value: %d", previousValue, newValue)
				errs = errs.Also(&apis.FieldError{Message: msg, Paths: []string{"UpdateRequests"}})
			}

			if spec.Source.NeedsUpdateRequestsIncrement(old.Spec.Source) {
				errs = errs.Also(&apis.FieldError{Message: "must increment UpdateRequests with change to source", Paths: []string{"UpdateRequests"}})
			}
		}
	}

	return errs
}

// Validate checks that the fields the user has specified in AppSpecInstances
// can be used together.
func (instances *AppSpecInstances) Validate(ctx context.Context) (errs *apis.FieldError) {
	hasExactly := instances.Exactly != nil
	hasMin := instances.Min != nil
	hasMax := instances.Max != nil

	if hasExactly && hasMin {
		errs = errs.Also(apis.ErrMultipleOneOf("exactly", "min"))
	}

	if hasExactly && hasMax {
		errs = errs.Also(apis.ErrMultipleOneOf("exactly", "max"))
	}

	if hasExactly && *instances.Exactly < 0 {
		errs = errs.Also(apis.ErrInvalidValue(*instances.Exactly, "exactly"))
	}

	if hasMin && *instances.Min < 0 {
		errs = errs.Also(apis.ErrInvalidValue(*instances.Min, "min"))
	}

	if hasMax && *instances.Max < 0 {
		errs = errs.Also(apis.ErrInvalidValue(*instances.Max, "max"))
	}

	if hasMin && hasMax && *instances.Min > *instances.Max {
		errs = errs.Also(&apis.FieldError{Message: "max must be >= min", Paths: []string{"min", "max"}})
	}

	return errs
}

// ValidatePodSpec proxies Knative Serving's checks on PodSpec, except for
// one condition. We don't allow setting the container image directly on the
// PodSpec because it'll be set by the source instead.
func ValidatePodSpec(podSpec v1.PodSpec) (errs *apis.FieldError) {
	// copy because we need to edit the PodSpec
	ps := podSpec.DeepCopy()

	switch len(ps.Containers) {
	case 0:
		errs = errs.Also(apis.ErrMissingField("containers"))
	case 1:
		if ps.Containers[0].Image != "" {
			errs = errs.Also(apis.ErrDisallowedFields("image"))
		}

		// Use a valid dummy image so we can re-use the validation from Knative
		// serving.
		ps.Containers[0].Image = "gcr.io/dummy/image:latest"
		errs = errs.Also(serving.ValidatePodSpec(*ps))
	default:
		errs = errs.Also(apis.ErrMultipleOneOf("containers"))
	}

	return errs
}

// ValidateServiceBindings validates each AppSpecServiceBinding for an App.
func (spec *AppSpec) ValidateServiceBindings(ctx context.Context) (errs *apis.FieldError) {
	for _, binding := range spec.ServiceBindings {
		errs = errs.Also(binding.Validate(ctx))
	}
	return errs
}

// Validate validates the fields of an AppSpecServiceBinding.
func (binding AppSpecServiceBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	if binding.BindingName == "" {
		errs = errs.Also(apis.ErrMissingField("bindingName"))
	}

	if binding.Instance == "" {
		errs = errs.Also(apis.ErrMissingField("instance"))
	}

	if !json.Valid(binding.Parameters) {
		errs = errs.Also(apis.ErrMissingField("parameters"))
	}

	return errs
}
