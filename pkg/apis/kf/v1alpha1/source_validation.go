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

	"knative.dev/pkg/apis"
)

// Validate checks for errors in the Source's spec or status fields.
func (source *Source) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if !apis.IsInStatusUpdate(ctx) {
		errs = errs.Also(source.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))
	}

	return errs
}

// Validate makes sure that a SourceSpec is properly configured.
func (spec *SourceSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	numDefined := countTrue(
		spec.IsContainerBuild(),
		spec.IsBuildpackBuild(),
		spec.IsDockerfileBuild(),
	)

	switch {
	case numDefined > 1:
		errs = errs.Also(apis.ErrMultipleOneOf("buildpackBuild", "containerImage", "dockerfile"))
	case numDefined == 0:
		errs = errs.Also(apis.ErrMissingOneOf("buildpackBuild", "containerImage", "dockerfile"))
	case spec.IsContainerBuild():
		errs = errs.Also(spec.ContainerImage.Validate(ctx))
	case spec.IsBuildpackBuild():
		errs = errs.Also(spec.BuildpackBuild.Validate(ctx))
	case spec.IsDockerfileBuild():
		errs = errs.Also(spec.Dockerfile.Validate(ctx))
	}

	return errs
}

func countTrue(vals ...bool) (count int) {
	for _, v := range vals {
		if v {
			count++
		}
	}

	return
}

// Validate makes sure that an SourceSpecContainerImage is properly configured.
func (containerImage *SourceSpecContainerImage) Validate(ctx context.Context) (errs *apis.FieldError) {

	if containerImage.Image == "" {
		errs = errs.Also(apis.ErrMissingField("image"))
	}

	return errs
}

// Validate makes sure that a SourceSpecBuildpackBuild is properly configured.
func (buildpackBuild *SourceSpecBuildpackBuild) Validate(ctx context.Context) (errs *apis.FieldError) {

	if buildpackBuild.Source == "" {
		errs = errs.Also(apis.ErrMissingField("source"))
	}

	if buildpackBuild.Stack == "" {
		errs = errs.Also(apis.ErrMissingField("stack"))
	}

	if buildpackBuild.BuildpackBuilder == "" {
		errs = errs.Also(apis.ErrMissingField("buildpackBuilder"))
	}

	if buildpackBuild.Image == "" {
		errs = errs.Also(apis.ErrMissingField("image"))
	}

	return errs
}

// Validate makes sure that a SourceSpecDockerfile is properly configured.
func (dockerfile *SourceSpecDockerfile) Validate(ctx context.Context) (errs *apis.FieldError) {
	if dockerfile.Image == "" {
		errs = errs.Also(apis.ErrMissingField("image"))
	}

	if dockerfile.Path == "" {
		errs = errs.Also(apis.ErrMissingField("path"))
	}

	if dockerfile.Source == "" {
		errs = errs.Also(apis.ErrMissingField("source"))
	}

	return errs
}
