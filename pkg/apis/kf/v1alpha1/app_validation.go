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
	"fmt"
	"math"

	"github.com/google/kf/v2/pkg/apis/kf"
	"knative.dev/pkg/apis"
)

// Validate checks for errors in the App's spec or status fields.
func (app *App) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(app.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(app.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate checks that the pod template the user has submitted is valid
// and that the scaling and lifecycle is valid.
func (spec *AppSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	errs = errs.Also(kf.ValidatePodSpec(spec.Template.Spec).ViaField("template.spec"))
	errs = errs.Also(spec.Instances.Validate(ctx).ViaField("instances"))
	errs = errs.Also(spec.Build.Validate(ctx).ViaField("build"))
	errs = errs.Also(spec.ValidateRoutes(ctx).ViaField("routes"))

	return errs
}

// ValidateBuildSpec validates the BuildSpec embedded in the AppSpec.
func (spec *AppSpecBuild) Validate(ctx context.Context) (errs *apis.FieldError) {

	// Ensure there is only one build method.
	{
		var methods int
		if spec.Spec != nil {
			methods++
		}
		if spec.Image != nil {
			methods++
		}
		if spec.BuildRef != nil {
			methods++
		}

		if methods == 0 {
			errs = errs.Also(apis.ErrMissingOneOf("spec", "image", "buildRef"))
		} else if methods > 1 {
			errs = errs.Also(apis.ErrMultipleOneOf("spec", "image", "buildRef"))
		}
	}

	// Ensure if ADX Builds is the method used, that the name is populated.
	if spec.BuildRef != nil && spec.BuildRef.Name == "" {
		errs = errs.Also(apis.ErrMissingField("buildRef.name"))
	}

	if spec.Spec != nil {
		errs = errs.Also(apis.CheckDisallowedFields(*spec.Spec, AppSpecBuildMask(*spec.Spec)))
		errs = errs.Also(spec.Spec.Validate(ctx).ViaField("spec"))
	}

	// Fail if the App Build has changed without changing the UpdateRequests.
	if base := apis.GetBaseline(ctx); base != nil {
		if old, ok := base.(*App); ok {
			previousValue := old.Spec.Build.UpdateRequests
			newValue := spec.UpdateRequests
			if previousValue > newValue {
				msg := fmt.Sprintf("UpdateRequests must be nondecreasing, previous value: %d new value: %d", previousValue, newValue)
				errs = errs.Also(&apis.FieldError{Message: msg, Paths: []string{"UpdateRequests"}})
			}

			if spec.NeedsUpdateRequestsIncrement(old.Spec.Build) {
				errs = errs.Also(&apis.FieldError{Message: "must increment UpdateRequests with change to build", Paths: []string{"UpdateRequests"}})
			}
		}
	}

	return errs
}

// Validate checks that the fields the user has specified in AppSpecInstances
// can be used together.
func (instances *AppSpecInstances) Validate(ctx context.Context) (errs *apis.FieldError) {
	hasReplicas := instances.Replicas != nil

	if hasReplicas && *instances.Replicas <= 0 {
		errs = errs.Also(apis.ErrInvalidValue(*instances.Replicas, "replicas"))
	}

	errs = errs.Also(instances.Autoscaling.Validate(ctx).ViaField("autoscaling"))

	return errs
}

// Validate checks that the fields the user has specified in AppSpecAutoscaling
// can be used together.
func (autoscaling *AppSpecAutoscaling) Validate(ctx context.Context) (errs *apis.FieldError) {
	if len(autoscaling.Rules) > 1 {
		errs = errs.Also(apis.ErrMultipleOneOf("rules"))
	}

	for idx, rule := range autoscaling.Rules {
		errs = errs.Also(rule.Validate(ctx).ViaFieldIndex("rules", idx))
	}

	minReplicas, maxReplicas := autoscaling.MinReplicas, autoscaling.MaxReplicas

	switch {
	case maxReplicas == nil && minReplicas == nil:
		// Autoscaling is not yet fully enabled, nothing to check.
	case maxReplicas == nil:
		errs = errs.Also(apis.ErrMissingField("maxReplicas"))
	case *maxReplicas <= 0:
		errs = errs.Also(apis.ErrOutOfBoundsValue(*autoscaling.MaxReplicas, 1, math.MaxInt32, "maxReplicas"))
	case minReplicas == nil:
		errs = errs.Also(apis.ErrMissingField("minReplicas"))
	case *minReplicas <= 0 || *minReplicas > *maxReplicas:
		errs = errs.Also(apis.ErrOutOfBoundsValue(*autoscaling.MinReplicas, 1, *maxReplicas, "minReplicas"))
	}

	return errs
}

// Validate checks that the fields the user has specified in AppAutoscalingRules
// can be used together.
func (r *AppAutoscalingRule) Validate(ctx context.Context) (errs *apis.FieldError) {
	target := r.Target

	switch {
	case r.RuleType != CPURuleType:
		errs = errs.Also(apis.ErrInvalidValue(r.RuleType, "ruleType"))
	case target == nil:
		errs = errs.Also(apis.ErrMissingField("target"))
	case *target <= 0 || *target > 100:
		// XXX: For different rule types, we may need impose different limits to Target.
		errs = errs.Also(apis.ErrOutOfBoundsValue(*r.Target, 1, 100, "target"))
	}

	return errs
}

// Validate implements Validatable.
func (s *Scale) Validate(ctx context.Context) (errs *apis.FieldError) {
	if s.Spec.Replicas < 0 {
		errs = errs.Also(apis.ErrInvalidValue(s.Spec.Replicas, "replicas"))
	}
	return errs
}

// ValidateRoutes validates each Route for an App.
func (spec *AppSpec) ValidateRoutes(ctx context.Context) (errs *apis.FieldError) {
	ctx = withAllowEmptyDomains(ctx)
	ctx = withAllowEmptyDestinationPort(ctx) // will be populated by the App

	for i := range spec.Routes {
		route := &spec.Routes[i]
		errs = errs.Also(route.Validate(ctx).ViaIndex(i))
	}

	return errs
}
