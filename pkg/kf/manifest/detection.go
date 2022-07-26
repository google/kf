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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
)

func errMultipleBuilds(buildNames ...string) error {
	sort.Strings(buildNames)
	csv := strings.Join(buildNames, ", ")
	return fmt.Errorf("app specifies multiple types of build, expected one: %s", csv)
}

// BuildSpecBuilder creates a BuildSpec.
type BuildSpecBuilder func(sourceImage string) (*v1alpha1.BuildSpec, error)

// DetectBuildType detects the correct BuildSpec for an Application.
// It also returns if a source image needs to be pushed.
func (app *Application) DetectBuildType(buildConfig v1alpha1.SpaceStatusBuildConfig) (BuildSpecBuilder, bool, error) {

	var matchedDetectors []buildTypeDetector
	for _, d := range defaultDetectors(buildConfig) {
		if d.CanBuild(*app) {
			matchedDetectors = append(matchedDetectors, d)
		}
	}

	if len(matchedDetectors) > 1 {
		var names []string
		for _, d := range matchedDetectors {
			names = append(names, d.Name())
		}
		return nil, false, errMultipleBuilds(names...)
	}

	if len(matchedDetectors) == 1 {
		return func(sourceImage string) (*v1alpha1.BuildSpec, error) {
			return matchedDetectors[0].CreateBuild(*app, sourceImage)
		}, matchedDetectors[0].RequiresSource(), nil
	}

	// If none of the detectors match, try defaulting the stack and trying againg
	wantV3 := buildConfig.DefaultToV3Stack
	hasV3 := len(buildConfig.StacksV3) > 0

	wantV2 := !wantV3
	hasV2 := len(buildConfig.StacksV2) > 0

	if hasV3 && (wantV3 || (wantV2 && !hasV2)) {
		appCopy := *app
		appCopy.Stack = buildConfig.StacksV3[0].Name

		detector := newBuildpackV3Detector(buildConfig)
		return func(sourceImage string) (*v1alpha1.BuildSpec, error) {
			return detector.CreateBuild(appCopy, sourceImage)
		}, detector.RequiresSource(), nil
	}

	if hasV2 {
		appCopy := *app
		appCopy.Stack = buildConfig.StacksV2[0].Name

		detector := newBuildpackV2Detector(buildConfig)
		return func(sourceImage string) (*v1alpha1.BuildSpec, error) {
			return detector.CreateBuild(appCopy, sourceImage)
		}, detector.RequiresSource(), nil
	}

	return nil, false, errors.New("can't detect the build type from the manifest")
}

func executeDetector(detector buildTypeDetector, app Application, sourceImage string) (*v1alpha1.BuildSpec, bool, error) {
	shouldPushSource := detector.RequiresSource()
	build, err := detector.CreateBuild(app, sourceImage)

	return build, shouldPushSource, err
}
