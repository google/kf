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

// Validate makes sure that ServiceInstance is properly configured.
func (instance *ServiceInstance) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(instance.GetObjectMeta()).ViaField("metadata"))

	// Deny changes to spec if the instance is a brokered service instance.
	if apis.IsInUpdate(ctx) && (instance.IsLegacyBrokered() || instance.IsKfBrokered()) {
		original := apis.GetBaseline(ctx).(*ServiceInstance)
		instance.Spec.DeleteRequests = original.Spec.DeleteRequests
		if diff, err := kmp.ShortDiff(original.Spec, instance.Spec); err != nil {
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
		errs = errs.Also(instance.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	}

	return errs
}

// Validate implements apis.Validatable.
func (spec *ServiceInstanceSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(spec.ServiceType.Validate(ctx))

	// Tags don't have any constraints according to the OSB API:
	// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#service-offering-object

	if spec.ParametersFrom.Name == "" {
		errs = errs.Also(apis.ErrMissingField("parametersFrom.name"))
	}

	return
}

// Validate implements apis.Validatable.
func (serviceType *ServiceType) Validate(ctx context.Context) (errs *apis.FieldError) {
	defined := sets.NewString()
	all := sets.NewString()

	fields := []struct {
		fieldName string
		isNil     bool
		validator apis.Validatable
	}{
		{
			fieldName: "userProvided",
			isNil:     serviceType.UPS == nil,
			validator: serviceType.UPS,
		},
		{
			fieldName: "brokered",
			isNil:     serviceType.Brokered == nil,
			validator: serviceType.Brokered,
		},
		{
			fieldName: "osb",
			isNil:     serviceType.OSB == nil,
			validator: serviceType.OSB,
		},
		{
			fieldName: "volume",
			isNil:     serviceType.Volume == nil,
			validator: serviceType.Volume,
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
func (instance *UPSInstance) Validate(ctx context.Context) (errs *apis.FieldError) {
	// Nothing to validate
	return
}

// Validate implements apis.Validatable.
func (instance *BrokeredInstance) Validate(ctx context.Context) (errs *apis.FieldError) {
	if instance.ClassName == "" {
		errs = errs.Also(apis.ErrMissingField("class"))
	}

	if instance.PlanName == "" {
		errs = errs.Also(apis.ErrMissingField("plan"))
	}

	return
}

// Validate implements apis.Validatable.
func (instance *OSBInstance) Validate(ctx context.Context) (errs *apis.FieldError) {
	for name, val := range map[string]string{
		"brokerName": instance.BrokerName,
		"classUID":   instance.ClassUID,
		"className":  instance.ClassName,
		"planUID":    instance.PlanUID,
		"planName":   instance.PlanName,
	} {
		if val == "" {
			errs = errs.Also(apis.ErrMissingField(name))
		}
	}

	// Test < 0 rather than <= 0 to ensure configurations prior to the field
	// addition are valid. The default function will upgrade those in the
	// reconciler.
	if v := instance.ProgressDeadlineSeconds; v < 0 {
		errs = errs.Also(apis.ErrOutOfBoundsValue(v, 1, math.MaxInt64, "progressDeadlineSeconds"))
	}

	return
}
