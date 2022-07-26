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

package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable.
func (sb *ServiceBroker) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(sb.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(sb.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return
}

// Validate implements apis.Validatable.
func (spec *ServiceBrokerSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(spec.CommonServiceBrokerSpec.Validate(ctx)) // no ViaField for embedded type

	if spec.Credentials.Name == "" {
		errs = errs.Also(apis.ErrMissingField("credentials.name"))
	}

	return
}

// Validate implements apis.Validatable.
func (sb *ClusterServiceBroker) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(sb.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(sb.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return
}

// Validate implements apis.Validatable.
func (spec *ClusterServiceBrokerSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(spec.CommonServiceBrokerSpec.Validate(ctx)) // no ViaField for embedded type

	if spec.VolumeBrokerSpec != nil {
		errs = errs.Also(spec.VolumeBrokerSpec.Validate(ctx))

		// skip Credentials validation for volume broker
		return
	}

	if spec.Credentials.Name == "" {
		errs = errs.Also(apis.ErrMissingField("credentials.name"))
	}

	if spec.Credentials.Namespace == "" {
		errs = errs.Also(apis.ErrMissingField("credentials.namespace"))
	}

	return
}

// Validate implements apis.Validatable.
func (sc *CommonServiceBrokerSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	// nothing to validate
	return
}

// Validate implements apis.Validatable.
func (spec *VolumeBrokerSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	if len(spec.VolumeOfferings) == 0 {
		errs = errs.Also(apis.ErrInvalidValue(spec.VolumeOfferings, "volumeOffering"))
	}

	return
}
