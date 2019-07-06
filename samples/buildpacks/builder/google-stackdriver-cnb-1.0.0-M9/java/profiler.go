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

package java

import (
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

// ProfilerDependency indicates that a JVM application should be run with Google Stackdriver Profiler enabled.
const ProfilerDependency = "google-stackdriver-profiler-java"

// Profiler represents an profiler contribution by the buildpack.
type Profiler struct {
	layer layers.DependencyLayer
}

// Contribute makes the contribution to launch.
func (p Profiler) Contribute() error {
	return p.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 0); err != nil {
			return err
		}

		// TODO: Default MODULE <extracted application name>
		// TODO: Default VERSION <extracted application version>
		return layer.WriteProfile("google-stackdriver-profiler", `if [[ -z "${BPL_GOOGLE_STACKDRIVER_MODULE+x}" ]]; then
    MODULE="default-module"
else
	MODULE=${BPL_GOOGLE_STACKDRIVER_MODULE}
fi

if [[ -z "${BPL_GOOGLE_STACKDRIVER_VERSION+x}" ]]; then
	VERSION=""
else
	VERSION=${BPL_GOOGLE_STACKDRIVER_VERSION}
fi

printf "Google Stackdriver Profiler enabled for ${MODULE}"

if [[ "${VERSION}" != "" ]]; then
	printf ":${VERSION}\n"
else
	printf "\n"
fi

AGENT="-agentpath:%s=--logtostderr=1,-cprof_service=${MODULE}"

if [[ "${VERSION}" != "" ]]; then
    AGENT="${AGENT},-cprof_service_version=${VERSION}"
fi

export JAVA_OPTS="${JAVA_OPTS} ${AGENT}"

`, filepath.Join(layer.Root, "profiler_java_agent.so"))
	}, layers.Launch)
}

// NewProfiler creates a new Profiler instance.
func NewProfiler(build build.Build) (Profiler, bool, error) {
	bp, ok := build.BuildPlan[ProfilerDependency]
	if !ok {
		return Profiler{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return Profiler{}, false, err
	}

	dep, err := deps.Best(ProfilerDependency, bp.Version, build.Stack)
	if err != nil {
		return Profiler{}, false, err
	}

	return Profiler{build.Layers.DependencyLayer(dep)}, true, nil
}
