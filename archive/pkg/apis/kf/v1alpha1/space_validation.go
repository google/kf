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

	"knative.dev/pkg/apis"
)

// Validate makes sure that Space is properly configured.
func (space *Space) Validate(ctx context.Context) (errs *apis.FieldError) {

	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	// validate name
	switch {
	case space.Name == "":
		errs = errs.Also(apis.ErrMissingField("name"))
	case space.Name == "kf" || space.Name == "default":
		errs = errs.Also(apis.ErrInvalidValue(space.Name, "name"))
	}

	errs = errs.Also(space.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate makes sure that SpaceSpec is properly configured.
func (s *SpaceSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(s.Security.Validate(ctx).ViaField("security"))
	errs = errs.Also(s.BuildpackBuild.Validate(ctx).ViaField("buildpackBuild"))
	errs = errs.Also(s.Execution.Validate(ctx).ViaField("execution"))
	errs = errs.Also(s.ResourceLimits.Validate(ctx).ViaField("resourceLimits"))

	return errs
}

// Validate makes sure that SpaceSpecSecurity is properly configured.
func (s *SpaceSpecSecurity) Validate(ctx context.Context) (errs *apis.FieldError) {
	// XXX: no validation
	return errs
}

// Validate makes sure that SpaceSpecBuildpackBuild is properly configured.
func (s *SpaceSpecBuildpackBuild) Validate(ctx context.Context) (errs *apis.FieldError) {
	if s.BuilderImage == "" {
		errs = errs.Also(apis.ErrMissingField("builderImage"))
	}

	if s.ContainerRegistry == "" {
		errs = errs.Also(apis.ErrMissingField("containerRegistry"))
	}

	return errs
}

// Validate makes sure that SpaceSpecExecution is properly configured.
func (s *SpaceSpecExecution) Validate(ctx context.Context) (errs *apis.FieldError) {
	if len(s.Domains) == 0 {
		return errs.Also(apis.ErrMissingField("domains"))
	}

	lastDefault := -1
	for i, d := range s.Domains {
		if !d.Default {
			continue
		}

		if lastDefault >= 0 {
			errs = errs.Also(
				&apis.FieldError{
					Paths:   []string{"domains"},
					Message: "multiple defaults",
					Details: "one domain must be set to default",
				},
			)
		}
		lastDefault = i
	}

	if lastDefault < 0 {
		errs = errs.Also(
			&apis.FieldError{
				Paths:   []string{"domains"},
				Message: "multiple defaults",
				Details: "one domain must be set to default",
			},
		)
	}

	return errs
}

// Validate makes sure that SpaceSpecResourceLimits is properly configured.
func (s *SpaceSpecResourceLimits) Validate(ctx context.Context) (errs *apis.FieldError) {
	// XXX: no validation
	return errs
}
