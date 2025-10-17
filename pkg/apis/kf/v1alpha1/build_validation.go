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
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

// Validate checks for errors in the Build's spec or status fields.
func (b *Build) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(b.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(b.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate makes sure that a BuildSpec is properly configured.
func (spec *BuildSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	if spec.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}

	validKinds := sets.NewString(
		string(tektonv1beta1.NamespacedTaskKind),
		BuiltinTaskKind,
	)

	// Check to ensure that a SOURCE_IMAGE is not set if a SourcePackage is
	// set.
	if spec.SourcePackage.Name != "" {
		for i, p := range spec.Params {
			if p.Name == SourceImageParamName {
				errs = errs.Also(apis.ErrInvalidArrayValue(p.Value, "params", i))
				break
			}
		}
	}

	if !validKinds.Has(spec.Kind) {
		errs = errs.Also(ErrInvalidEnumValue(spec.Kind, "kind", validKinds.List()))
	}

	cfg, err := config.FromContext(ctx).Defaults()

	switch {
	case err != nil:
		errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to load config: %q", err)))
	case cfg.FeatureFlags.DisableCustomBuildsFlag().IsEnabled() && spec.isCustomBuild(ctx):
		errs = errs.Also(apis.ErrGeneric(
			fmt.Sprintf("Custom Builds are disabled, kind must be %q but was %q", BuiltinTaskKind, spec.Kind), "kind"))
	case cfg.FeatureFlags.DockerfileBuilds().IsDisabled() && spec.isDockerfileBuild():
		errs = errs.Also(apis.ErrGeneric(
			fmt.Sprintf(
				"Dockerfile Builds are disabled, but BuildTaskRef name was %q", DockerfileBuildTaskName), "name"))
	case cfg.FeatureFlags.CustomBuildpacks().IsDisabled() && spec.isCustomBuildpackV2Build(cfg):
		errs = errs.Also(apis.ErrGeneric(
			"Builds are restricted to configured Buildpacks, but Build depends on other Buildpacks", "name"))
	case cfg.FeatureFlags.CustomBuildpacks().IsDisabled() && spec.isCustomV3Build(cfg):
		errs = errs.Also(apis.ErrGeneric(
			"Builds are restricted to configured Buildpacks, but Build potentially depends on other Buildpacks through a custom stack", "name"))
	case cfg.FeatureFlags.CustomStacks().IsDisabled() && spec.isCustomStackV2Build(cfg) ||
		cfg.FeatureFlags.CustomStacks().IsDisabled() && spec.isCustomV3Build(cfg):
		errs = errs.Also(apis.ErrGeneric(
			"Builds are restricted to configured Stacks, but Build depends on other Stacks", "name"))
	}
	return errs
}

// Returns true if spec meets the criteria of a custom Build that should be blocked.
func (spec *BuildSpec) isCustomBuild(ctx context.Context) bool {
	return apis.IsInCreate(ctx) && spec.Kind != BuiltinTaskKind
}

// Returns true if the Build matches the supplied type.
func (spec *BuildSpec) isBuildType(buildType string) bool {
	return spec.Kind == BuiltinTaskKind && spec.BuildTaskRef.Name == buildType
}

// Returns true if spec is a Dockerfile Build.
func (spec *BuildSpec) isDockerfileBuild() bool {
	return spec.isBuildType(DockerfileBuildTaskName)
}

// Returns true if the BuildSpec includes a Buildpack that is not configured in the list
// of V2 Buildpacks for the Space when the feature flag is enabled.
func (spec *BuildSpec) isCustomBuildpackV2Build(cfg *config.DefaultsConfig) bool {
	if !spec.isBuildType(BuildpackV2BuildTaskName) {
		return false
	}
	packs := ""
	found := false
	for _, x := range spec.Params {
		if x.Name == BuildpackV2ParamName {
			packs = x.Value
			found = true
		}
	}
	if !found {
		return true
	}
	// Short circuit return to avoid hassle of empty string element being considered a custom Buildpack
	if packs == "" {
		return false
	}

	bps := strings.Split(packs, ",")
	spacebps := sets.NewString()
	for _, x := range cfg.SpaceBuildpacksV2 {
		spacebps.Insert(x.Name)
	}
	return !spacebps.HasAll(bps...)
}

// Returns true if spec is a Buildpack V2 BuildSpec that uses a Stack that is not
// included in the list of V2 Stacks configured on the Space.
func (spec *BuildSpec) isCustomStackV2Build(cfg *config.DefaultsConfig) bool {
	if !spec.isBuildType(BuildpackV2BuildTaskName) {
		return false
	}
	stack := ""
	found := false
	for _, x := range spec.Env {
		if x.Name == StackV2EnvVarName {
			stack = x.Value
			found = true
		}
	}
	if !found {
		return true
	}
	return cfg.SpaceStacksV2.FindStackByName(stack) == nil
}

// Returns true if the Build is a V3 Buildpack Build that specifies a custom Stack outside
// of the V3 Stacks configured on the Space. It is not possible to specify a custom V3
// Buildpack unless a custom Stack is provided to the push command.
func (spec *BuildSpec) isCustomV3Build(cfg *config.DefaultsConfig) bool {
	if spec.Kind != BuiltinTaskKind || spec.BuildTaskRef.Name != BuildpackV3BuildTaskName {
		return false
	}
	return spec.isCustomStackV3(cfg)
}

// Returns true if the Stack for the Build is outside of the list of Space stacks
func (spec *BuildSpec) isCustomStackV3(cfg *config.DefaultsConfig) bool {
	stackImage := ""
	found := false
	for _, x := range spec.Params {
		if x.Name == RunImageParamName {
			stackImage = x.Value
			found = true
		}
	}
	if !found {
		return true
	}
	for _, x := range cfg.SpaceStacksV3 {
		if stackImage == x.RunImage {
			return false
		}
	}
	return true
}
