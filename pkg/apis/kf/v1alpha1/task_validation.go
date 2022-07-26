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

	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
)

// Validate makes sure that Task is properly configured.
func (t *Task) Validate(ctx context.Context) (errs *apis.FieldError) {
	if apis.IsInStatusUpdate(ctx) {
		return
	}
	errs = errs.Also(apis.ValidateObjectMetadata(t.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(t.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	return
}

// Validate implements apis.Validatable.
func (spec *TaskSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	if spec.AppRef.Name == "" {
		errs = errs.Also(apis.ErrMissingField("appRef"))
	}

	if len(spec.CPU) > 0 {
		if _, err := resource.ParseQuantity(spec.CPU); err != nil {
			errs = errs.Also(apis.ErrInvalidValue(spec.CPU, "cpu"))
		}
	}

	if len(spec.Memory) > 0 {
		if _, err := resource.ParseQuantity(spec.Memory); err != nil {
			errs = errs.Also(apis.ErrInvalidValue(spec.Memory, "memory"))
		}
	}

	if len(spec.Disk) > 0 {
		if _, err := resource.ParseQuantity(spec.Disk); err != nil {
			errs = errs.Also(apis.ErrInvalidValue(spec.Disk, "disk"))
		}
	}

	return errs
}
