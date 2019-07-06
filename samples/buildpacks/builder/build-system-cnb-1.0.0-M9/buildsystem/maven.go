/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buildsystem

import (
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/openjdk-cnb/jdk"
)

// MavenDependency is the key identifying the Maven build system in the buildpack plan.
const MavenDependency = "maven"

// MavenBuildPlanContribution returns the BuildPlan with requirements for Maven.
func MavenBuildPlanContribution(buildPlan buildplan.BuildPlan) buildplan.BuildPlan {
	return buildplan.BuildPlan{
		MavenDependency:           buildPlan[MavenDependency],
		jvmapplication.Dependency: buildPlan[jvmapplication.Dependency],
		jdk.Dependency:            buildPlan[jdk.Dependency],
	}
}

// IsMaven returns whether this application is built using Maven.
func IsMaven(application application.Application) bool {
	exists, err := helper.FileExists(filepath.Join(application.Root, "pom.xml"))
	if err != nil {
		return false
	}

	return exists
}

// NewMavenBuildSystem creates a new Maven BuildSystem instance. OK is true if build plan contains "maven" dependency,
// otherwise false.
func NewMavenBuildSystem(build build.Build) (BuildSystem, bool, error) {
	bp, ok := build.BuildPlan[MavenDependency]
	if !ok {
		return BuildSystem{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return BuildSystem{}, false, err
	}

	dep, err := deps.Best(MavenDependency, bp.Version, build.Stack)
	if err != nil {
		return BuildSystem{}, false, err
	}

	layer := build.Layers.DependencyLayer(dep)
	distribution := filepath.Join(layer.Root, "bin", "mvn")
	wrapper := filepath.Join(build.Application.Root, "mvnw")

	return BuildSystem{
		contributeMavenDistribution,
		distribution,
		layer,
		build.Logger,
		wrapper,
	}, true, nil
}

func contributeMavenDistribution(artifact string, layer layers.DependencyLayer) error {
	layer.Logger.Body("Expanding to %s", layer.Root)
	return helper.ExtractTarGz(artifact, layer.Root, 1)
}
