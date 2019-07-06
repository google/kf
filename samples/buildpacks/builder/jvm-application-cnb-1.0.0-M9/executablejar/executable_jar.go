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

package executablejar

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/mitchellh/mapstructure"
)

// Dependency indicates that an application is an executable JAR.
const Dependency = "executable-jar"

// ExecutableJAR represents an executable JAR JVM application.
type ExecutableJAR struct {
	// Metadata is metadata about the executable JAR application.
	Metadata Metadata

	layer  layers.Layer
	layers layers.Layers
}

// Contribute makes the contribution to launch.
func (e ExecutableJAR) Contribute() error {
	if err := e.layer.Contribute(e.Metadata, func(layer layers.Layer) error {
		return layer.AppendPathSharedEnv("CLASSPATH", strings.Join(e.Metadata.ClassPath, string(filepath.ListSeparator)))
	}, layers.Build, layers.Cache, layers.Launch); err != nil {
		return err
	}

	command := fmt.Sprintf("java -cp $CLASSPATH $JAVA_OPTS %s", e.Metadata.MainClass)

	return e.layers.WriteApplicationMetadata(layers.Metadata{
		Processes: layers.Processes{
			{"executable-jar", command},
			{"task", command},
			{"web", command},
		},
	})
}

// BuildPlan returns the dependency information for this application.
func (e ExecutableJAR) BuildPlan() (buildplan.BuildPlan, error) {
	md := make(buildplan.Metadata)

	if err := mapstructure.Decode(e.Metadata, &md); err != nil {
		return nil, err
	}

	return buildplan.BuildPlan{Dependency: buildplan.Dependency{Metadata: md}}, nil
}

// NewExecutableJAR creates a new ExecutableJAR instance.  OK is true if the build plan contains a "jvm-application"
// dependency and a "Main-Class" manifest key.
func NewExecutableJAR(build build.Build) (ExecutableJAR, bool, error) {
	_, ok := build.BuildPlan[jvmapplication.Dependency]
	if !ok {
		return ExecutableJAR{}, false, nil
	}

	md, ok, err := NewMetadata(build.Application, build.Logger)
	if err != nil {
		return ExecutableJAR{}, false, err
	}

	if !ok {
		return ExecutableJAR{}, false, nil
	}

	return ExecutableJAR{
		md,
		build.Layers.Layer(Dependency),
		build.Layers,
	}, true, nil
}
