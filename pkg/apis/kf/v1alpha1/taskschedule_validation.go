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

	cron "github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

// Validate makes sure that TaskSchedule is properly configured.
func (t *TaskSchedule) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}
	errs = errs.Also(apis.ValidateObjectMetadata(t.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(t.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	return
}

var validConcurrencyPolicies = sets.NewString(ConcurrencyPolicyAlways, ConcurrencyPolicyForbid, ConcurrencyPolicyReplace)

// Validate implements apis.Validatable.
func (spec *TaskScheduleSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	if _, err := cron.ParseStandard(spec.Schedule); err != nil {
		errs = errs.Also(apis.ErrInvalidValue(spec.Schedule, "schedule"))
	}

	if !validConcurrencyPolicies.Has(spec.ConcurrencyPolicy) {
		errs = errs.Also(apis.ErrInvalidValue(spec.ConcurrencyPolicy, "concurrencyPolicy"))
	}

	errs = errs.Also(spec.TaskTemplate.Validate(ctx).ViaField("taskTemplate"))
	return
}
