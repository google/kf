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

package jre

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/openjdk-cnb/internal"
)

const (
	// BuildContribution is a build plan dependency key indicating a requirement for the dependency at build time.
	BuildContribution string = "build"

	// Dependency is a build plan dependency indicating a requirement for a JRE.
	Dependency string = "openjdk-jre"

	// LaunchContribution is a build plan dependency yet indicate a requirement for the dependency at launch time.
	LaunchContribution string = "launch"
)

// JRE represents a JRE contribution by the buildpack.
type JRE struct {
	buildContribution  bool
	layer              layers.DependencyLayer
	launchContribution bool
}

// Contribute contributes an expanded JRE to a cache layer.
func (j JRE) Contribute() error {
	return j.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("JAVA_HOME", layer.Root); err != nil {
			return err
		}

		return nil
	}, j.flags()...)
}

func (j JRE) flags() []layers.Flag {
	var flags []layers.Flag

	if j.buildContribution {
		flags = append(flags, layers.Build, layers.Cache)
	}

	if j.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}

// NewJRE creates a new JRE instance. OK is true if build plan contains "openjdk-jre" dependency, otherwise false.
func NewJRE(build build.Build) (JRE, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return JRE{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return JRE{}, false, err
	}

	version, err := internal.Version(Dependency, bp, build.Buildpack)
	if err != nil {
		return JRE{}, false, err
	}

	dep, err := deps.Best(Dependency, version, build.Stack)
	if err != nil {
		return JRE{}, false, err
	}

	jre := JRE{layer: build.Layers.DependencyLayer(dep)}

	if _, ok := bp.Metadata[BuildContribution]; ok {
		jre.buildContribution = true
	}

	if _, ok := bp.Metadata[LaunchContribution]; ok {
		jre.launchContribution = true
	}

	return jre, true, nil
}
