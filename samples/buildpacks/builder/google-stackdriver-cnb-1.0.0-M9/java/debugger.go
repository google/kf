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

// DebuggerDependency indicates that a JVM application should be run with Google Stackdriver Debugger enabled.
const DebuggerDependency = "google-stackdriver-debugger-java"

// Debugger represents an debugger contribution by the buildpack.
type Debugger struct {
	layer layers.DependencyLayer
}

// Contribute makes the contribution to launch.
func (d Debugger) Contribute() error {
	return d.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 0); err != nil {
			return err
		}

		// TODO: Default MODULE <extracted application name>
		// TODO: Default VERSION <extracted application version>
		return layer.WriteProfile("google-stackdriver-debugger", `if [[ -z "${BPL_GOOGLE_STACKDRIVER_MODULE+x}" ]]; then
    MODULE="default-module"
else
	MODULE=${BPL_GOOGLE_STACKDRIVER_MODULE}
fi

if [[ -z "${BPL_GOOGLE_STACKDRIVER_VERSION+x}" ]]; then
	VERSION=""
else
	VERSION=${BPL_GOOGLE_STACKDRIVER_VERSION}
fi

printf "Google Stackdriver Debugger enabled for ${MODULE}"

if [[ "${VERSION}" != "" ]]; then
	printf ":${VERSION}\n"
else
	printf "\n"
fi

export JAVA_OPTS="${JAVA_OPTS} -agentpath:%s=--logtostderr=1 -Dcom.google.cdbg.auth.serviceaccount.enable=true -Dcom.google.cdbg.module=${MODULE}"

if [[ "${VERSION}" != "" ]]; then
    export JAVA_OPTS="${JAVA_OPTS} -Dcom.google.cdbg.version=${VERSION}"
fi
`, filepath.Join(layer.Root, "cdbg_java_agent.so"))
	}, layers.Launch)
}

// NewDebugger creates a new Debugger instance.
func NewDebugger(build build.Build) (Debugger, bool, error) {
	bp, ok := build.BuildPlan[DebuggerDependency]
	if !ok {
		return Debugger{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return Debugger{}, false, err
	}

	dep, err := deps.Best(DebuggerDependency, bp.Version, build.Stack)
	if err != nil {
		return Debugger{}, false, err
	}

	return Debugger{build.Layers.DependencyLayer(dep)}, true, nil
}
