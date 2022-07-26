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

package manifest

import (
	"fmt"
	"net/url"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
)

// buildTypeDetector is a definition for detecting from an Application manifest
// if a given Tekton Task type should be used for building the application.
type buildTypeDetector interface {
	// Name is a human-readable name for the build type
	Name() string

	// CanBuild returns true if the given application can be built.
	CanBuild(application Application) bool

	// RequiresSource signals if a source image will need to be pushed.
	RequiresSource() bool

	// CreateBuild returns a BuildSpec for the given application.
	CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error)
}

type dockerDetector struct{}

func (*dockerDetector) Name() string {
	return "Dockerfile"
}

func (*dockerDetector) CanBuild(application Application) bool {
	return application.Dockerfile.Path != ""
}

func (*dockerDetector) RequiresSource() bool {
	return true
}

func (*dockerDetector) CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error) {
	build := v1alpha1.DockerfileBuild(sourceImage, application.Dockerfile.Path)
	return &build, nil
}

// newDockerfileDetector creates a detector for Dockerfiles.
func newDockerfileDetector() buildTypeDetector {
	return &dockerDetector{}
}

type buildpackV3Detector struct {
	knownStacks config.StackV3List
}

func (*buildpackV3Detector) Name() string {
	return "Buildpack V3"
}

func (d *buildpackV3Detector) getStack(stackName string) *config.StackV3Definition {
	for _, v := range d.knownStacks {
		if stackName == v.Name {
			return &v
		}
	}

	return nil
}

func (d *buildpackV3Detector) CanBuild(application Application) bool {
	return d.getStack(application.Stack) != nil
}

func (*buildpackV3Detector) RequiresSource() bool {
	return true
}

func (d *buildpackV3Detector) CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error) {
	stack := d.getStack(application.Stack)
	if stack == nil {
		return nil, fmt.Errorf("couldn't find stack %s", application.Stack)
	}

	build := v1alpha1.BuildpackV3Build(sourceImage, *stack, application.BuildpacksSlice())
	return &build, nil
}

// newBuildpackV3Detector detects if apps can be built with Cloud Native Buildpacks.
func newBuildpackV3Detector(buildConfig v1alpha1.SpaceStatusBuildConfig) buildTypeDetector {
	return &buildpackV3Detector{
		knownStacks: buildConfig.StacksV3,
	}
}

type buildpackV2Detector struct {
	knownStacks     config.StackV2List
	knownBuildpacks config.BuildpackV2List
}

func (*buildpackV2Detector) Name() string {
	return "Buildpack V2"
}

// looksLikeBuildpackV2URL checks if the given string is a "valid" buildpack
// by checking that it is a URL and has a scheme. Most things in go will pass
// URL parsing, but we know buildpack URLs must at least have a scheme.
func looksLikeBuildpackV2URL(buildpackString string) bool {
	parsed, err := url.Parse(buildpackString)
	if err != nil || parsed.Scheme == "" {
		return false
	}

	return true
}

func (d *buildpackV2Detector) getStack(stackName string) *config.StackV2Definition {
	for _, v := range d.knownStacks {
		if stackName == v.Name {
			return &v
		}
	}

	return nil
}

func (d *buildpackV2Detector) CanBuild(application Application) bool {
	if d.getStack(application.Stack) == nil {
		return false
	}

	return true
}

func (*buildpackV2Detector) RequiresSource() bool {
	return true
}

func (d *buildpackV2Detector) defaultBuildpacks() (out []string) {
	for _, b := range d.knownBuildpacks {
		out = append(out, b.URL)
	}

	return
}

func (d *buildpackV2Detector) CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error) {
	stack := d.getStack(application.Stack)
	if stack == nil {
		return nil, fmt.Errorf("couldn't find stack %s", application.Stack)
	}

	// If the user has explicitly specified buildpacks they want all of them
	// to be run in the order specified. This allows users to override orders
	// and/or enable the use of buildpacks that always fail the detect step.
	// For example, a user could want to use the node buildpack to build their
	// website but the staticfile buildpack to serve it.
	buildpacks := application.BuildpacksSlice()
	skipDetect := true
	if len(buildpacks) == 0 {
		buildpacks = d.defaultBuildpacks()
		skipDetect = false
	} else {
		for i, orig := range buildpacks {
			if !looksLikeBuildpackV2URL(orig) {
				buildpacks[i] = d.knownBuildpacks.MapToURL(orig)
			}
		}
	}

	build := v1alpha1.BuildpackV2Build(sourceImage, *stack, buildpacks, skipDetect)
	return &build, nil
}

// newBuildpackV2Detector detects if apps can be built with CloudFoundry Buildpacks.
func newBuildpackV2Detector(buildConfig v1alpha1.SpaceStatusBuildConfig) buildTypeDetector {
	return &buildpackV2Detector{
		knownStacks:     buildConfig.StacksV2,
		knownBuildpacks: buildConfig.BuildpacksV2,
	}
}

type customBuildDetector struct{}

func (*customBuildDetector) Name() string {
	return "Custom TaskRun"
}

func (c *customBuildDetector) CanBuild(application Application) bool {
	return application.KfApplicationExtension.Build != nil
}

func (*customBuildDetector) RequiresSource() bool {
	return true
}

func (c *customBuildDetector) CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error) {
	buildCopy := *application.KfApplicationExtension.Build

	sourceParam := v1alpha1.StringParam(v1alpha1.SourceImageParamName, sourceImage)
	foundSource := false
	for i, param := range buildCopy.Params {
		if param.Name == v1alpha1.SourceImageParamName {
			buildCopy.Params[i] = sourceParam
			foundSource = true
		}
	}

	if !foundSource {
		buildCopy.Params = append(buildCopy.Params, sourceParam)
	}

	return &buildCopy, nil
}

// newCustomBuildDetector uses the app's build configuration.
func newCustomBuildDetector() buildTypeDetector {
	return &customBuildDetector{}
}

type dockerImageDetector struct{}

func (*dockerImageDetector) Name() string {
	return "Docker Image"
}

func (*dockerImageDetector) CanBuild(application Application) bool {
	return application.Docker.Image != ""
}

func (*dockerImageDetector) RequiresSource() bool {
	return false
}

func (*dockerImageDetector) CreateBuild(application Application, sourceImage string) (*v1alpha1.BuildSpec, error) {
	return nil, nil
}

// newDockerImageDetector detects if apps can skip the build step because they
// reference a pre-built container image.
func newDockerImageDetector() buildTypeDetector {
	return &dockerImageDetector{}
}

// defaultDetectors is the default ordered list of BuildTypeDetector used in Kf.
func defaultDetectors(buildConfig v1alpha1.SpaceStatusBuildConfig) []buildTypeDetector {
	return []buildTypeDetector{
		newDockerImageDetector(),
		newCustomBuildDetector(),
		newDockerfileDetector(),
		newBuildpackV2Detector(buildConfig),
		newBuildpackV3Detector(buildConfig),
	}
}
