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

package kf

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

const (
	routeServicesFlag       = "enable_route_services"
	disableCustomBuildsFlag = "disable_custom_builds"
	dockerfileBuildsFlag    = "enable_dockerfile_builds"
	customBuildpacksFlag    = "enable_custom_buildpacks"
	customStacksFlag        = "enable_custom_stacks"
)

// BuildpackV2List holds an array of BuildpackV2Definition.
// Its primary use is doing validation over a list of Buildpacks.
type BuildpackV2List []BuildpackV2Definition

var _ apis.Validatable = (BuildpackV2List)(nil)

// Validate implements apis.Validatable
func (list BuildpackV2List) Validate(ctx context.Context) (err *apis.FieldError) {
	names := sets.NewString()

	for idx, val := range list {
		if names.Has(val.Name) {
			err = err.Also((&apis.FieldError{
				Message: "duplicate name",
				Details: fmt.Sprintf("the name %q is duplicated", val.Name),
				Paths:   []string{"name"},
			}).ViaIndex(idx))
		}

		err = err.Also(val.Validate(ctx).ViaIndex(idx))

		names.Insert(val.Name)
	}

	return
}

// WithoutDisabled returns a list of buildpacks with the disabled ones filtered
// out.
func (list BuildpackV2List) WithoutDisabled() BuildpackV2List {
	var out BuildpackV2List

	for _, buildpack := range list {
		if !buildpack.Disabled {
			out = append(out, buildpack)
		}
	}

	return out
}

// MapToURL looks up the buildpack in the given list and returns the URL of the
// corresponding buildpack. If the name isn't found, it's returned directly
// presuming it's either built-in or a URL already.
func (list BuildpackV2List) MapToURL(name string) string {
	for _, buildpack := range list {
		if buildpack.Name == name {
			return buildpack.URL
		}
	}

	return name
}

// BuildpackV2Definition contains the a definition of a buildpack.
type BuildpackV2Definition struct {
	// Name is the human readable name of the buildpack.
	Name string `json:"name"`
	// URL is the URL of the given buildpack.
	URL string `json:"url"`
	// Disabled is set when this buildpack shouldn't be used to
	// build apps.
	Disabled bool `json:"disabled"`
}

// Validate implements apis.Validatable
func (defn *BuildpackV2Definition) Validate(ctx context.Context) (err *apis.FieldError) {
	if defn.Name == "" {
		err = err.Also(apis.ErrMissingField("name"))
	}

	if defn.URL == "" {
		err = err.Also(apis.ErrMissingField("url"))
	}

	return
}

// StackV2List holds an array of StackV2Definition.
// Its primary use is doing validation over a list of Stacks.
type StackV2List []StackV2Definition

var _ apis.Validatable = (StackV2List)(nil)

// Validate implements apis.Validatable
func (list StackV2List) Validate(ctx context.Context) (err *apis.FieldError) {
	names := sets.NewString()

	for idx, val := range list {
		if names.Has(val.Name) {
			err = err.Also((&apis.FieldError{
				Message: "duplicate name",
				Details: fmt.Sprintf("the name %q is duplicated", val.Name),
				Paths:   []string{"name"},
			}).ViaIndex(idx))
		}

		err = err.Also(val.Validate(ctx).ViaIndex(idx))

		names.Insert(val.Name)
	}

	return
}

// FindStackByName returns the first stack with the given name or nil if none
// exists.
func (list StackV2List) FindStackByName(name string) *StackV2Definition {
	for _, v := range list {
		if v.Name == name {
			return &v
		}
	}

	return nil
}

// StackV2Definition contains the definition of a stack.
type StackV2Definition struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Description  string            `json:"description,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// Validate implements apis.Validatable
func (defn *StackV2Definition) Validate(ctx context.Context) (err *apis.FieldError) {
	if defn.Name == "" {
		err = err.Also(apis.ErrMissingField("name"))
	}

	if defn.Image == "" {
		err = err.Also(apis.ErrMissingField("image"))
	}

	return
}

// StackV3List holds an array of StackV3Definition.
// Its primary use is doing validation over a list of Stacks.
type StackV3List []StackV3Definition

var _ apis.Validatable = (StackV3List)(nil)

// Validate implements apis.Validatable
func (list StackV3List) Validate(ctx context.Context) (err *apis.FieldError) {
	names := sets.NewString()

	for idx, val := range list {
		if names.Has(val.Name) {
			err = err.Also((&apis.FieldError{
				Message: "duplicate name",
				Details: fmt.Sprintf("the name %q is duplicated", val.Name),
				Paths:   []string{"name"},
			}).ViaIndex(idx))
		}

		err = err.Also(val.Validate(ctx).ViaIndex(idx))

		names.Insert(val.Name)
	}

	return
}

// StackV3Definition contains the definition of a cloud native buildpack stack.
type StackV3Definition struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	BuildImage   string            `json:"buildImage"`
	RunImage     string            `json:"runImage"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// Validate implements apis.Validatable
func (defn *StackV3Definition) Validate(ctx context.Context) (err *apis.FieldError) {
	if defn.Name == "" {
		err = err.Also(apis.ErrMissingField("name"))
	}

	if defn.BuildImage == "" {
		err = err.Also(apis.ErrMissingField("buildImage"))
	}

	if defn.RunImage == "" {
		err = err.Also(apis.ErrMissingField("runImage"))
	}

	return
}

// DomainTemplate mimics the structure of v1alpha1.SpaceDomain
type DomainTemplate struct {
	// Domain is the valid domain that can be used in conjunction with a
	// hostname and path for a route.
	Domain string `json:"domain"`
	// GatewayName is the name of the Istio Gateway supported by the domain.
	// Values can include a Namespace as a prefix.
	// Only the kf Namespace is allowed e.g. kf/some-gateway.
	// See https://istio.io/docs/reference/config/networking/gateway/
	GatewayName string `json:"gatewayName,omitempty"`
}

// FeatureFlagToggles maps a feature name to a bool representing whether the feature is enabled.
type FeatureFlagToggles map[string]bool

// FeatureFlag is the internal representation of a feature flag.
// If a feature flag is not set in the config-defaults ConfigMap, the value of the flag internally is still defaulted.
// +k8s:deepcopy-gen=false
type FeatureFlag struct {
	Name      string `json:"name"`
	Default   bool   `json:"default"`
	isEnabled func(*FeatureFlag) bool
}

// IsEnabled returns whether or not the feature is enabled.
func (ff *FeatureFlag) IsEnabled() bool {
	return ff.isEnabled(ff)
}

// IsDisabled returns whether or not the feature is disabled.
func (ff *FeatureFlag) IsDisabled() bool {
	return !ff.isEnabled(ff)
}

// RouteServices returns the FeatureFlag for enabling route services.
func (fft FeatureFlagToggles) RouteServices() *FeatureFlag {
	return &FeatureFlag{
		Name:    routeServicesFlag,
		Default: false,
		isEnabled: func(ff *FeatureFlag) bool {
			if setValue, ok := fft[routeServicesFlag]; ok {
				return setValue
			}
			return ff.Default
		},
	}
}

// SetRouteServices sets the value for the route services feature flag on a FeatureFlagToggles map.
func (fft FeatureFlagToggles) SetRouteServices(toggle bool) {
	fft[routeServicesFlag] = toggle
}

// DisableCustomBuildsFlag returns the FeatureFlag for disabling custom builds.
func (fft FeatureFlagToggles) DisableCustomBuildsFlag() *FeatureFlag {
	return &FeatureFlag{
		Name:    disableCustomBuildsFlag,
		Default: false,
		isEnabled: func(ff *FeatureFlag) bool {
			if setValue, ok := fft[disableCustomBuildsFlag]; ok {
				return setValue
			}
			return ff.Default
		},
	}
}

// SetDisableCustomBuilds sets the value for the custom build disabling feature flag on a FeatureFlagToggles map.
func (fft FeatureFlagToggles) SetDisableCustomBuilds(disabled bool) {
	fft[disableCustomBuildsFlag] = disabled
}

// DockerfileBuilds returns the FeatureFlag for enabling docker image based builds.
func (fft FeatureFlagToggles) DockerfileBuilds() *FeatureFlag {
	return &FeatureFlag{
		Name:    dockerfileBuildsFlag,
		Default: true,
		isEnabled: func(ff *FeatureFlag) bool {
			if setValue, ok := fft[dockerfileBuildsFlag]; ok {
				return setValue
			}
			return ff.Default
		},
	}
}

// SetDockerfileBuilds sets the value for the docker build feature flag on a FeatureFlagToggles map.
func (fft FeatureFlagToggles) SetDockerfileBuilds(enabled bool) {
	fft[dockerfileBuildsFlag] = enabled
}

// CustomBuildpacks returns the FeatureFlag for enabling buildpacks outside of the Space-configured ones,
func (fft FeatureFlagToggles) CustomBuildpacks() *FeatureFlag {
	return &FeatureFlag{
		Name:    customBuildpacksFlag,
		Default: true,
		isEnabled: func(ff *FeatureFlag) bool {
			if setValue, ok := fft[customBuildpacksFlag]; ok {
				return setValue
			}
			return ff.Default
		},
	}
}

// SetCustomBuildpacks sets the value for the custom buildpack feature flag on a FeatureFlagToggles map.
func (fft FeatureFlagToggles) SetCustomBuildpacks(enabled bool) {
	fft[customBuildpacksFlag] = enabled
}

// CustomStacks returns the FeatureFlag for enabling stacks outside of the Space-configured ones.,
func (fft FeatureFlagToggles) CustomStacks() *FeatureFlag {
	return &FeatureFlag{
		Name:    customStacksFlag,
		Default: true,
		isEnabled: func(ff *FeatureFlag) bool {
			if setValue, ok := fft[customStacksFlag]; ok {
				return setValue
			}
			return ff.Default
		},
	}
}

// SetCustomStacks sets the value for the custom stack feature flag on a FeatureFlagToggles map.
func (fft FeatureFlagToggles) SetCustomStacks(enabled bool) {
	fft[customStacksFlag] = enabled
}

// BuildNodeSelectors maps node selector keys with node selector values.
type BuildNodeSelectors map[string]string
