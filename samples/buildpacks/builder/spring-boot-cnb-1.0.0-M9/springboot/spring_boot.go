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

package springboot

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

// Dependency indicates that an application is a Spring Boot application.
const Dependency = "spring-boot"

// SpringBoot represents a Spring Boot JVM application.
type SpringBoot struct {
	// Metadata is metadata about the Spring Boot application.
	Metadata Metadata

	layer  layers.Layer
	layers layers.Layers
}

// Contribute makes the contribution to build, cache, and launch.
func (s SpringBoot) Contribute() error {
	if err := s.layer.Contribute(s.Metadata, func(layer layers.Layer) error {
		return layer.AppendPathSharedEnv("CLASSPATH", strings.Join(s.Metadata.ClassPath, string(filepath.ListSeparator)))
	}, layers.Build, layers.Cache, layers.Launch); err != nil {
		return err
	}

	command := fmt.Sprintf("java -cp $CLASSPATH $JAVA_OPTS %s", s.Metadata.StartClass)

	return s.layers.WriteApplicationMetadata(layers.Metadata{
		Processes: layers.Processes{
			{"spring-boot", command},
			{"task", command},
			{"web", command},
		},
	})
}

// BuildPlan returns the dependency information for this application.
func (s SpringBoot) BuildPlan() (buildplan.BuildPlan, error) {
	md := make(buildplan.Metadata)

	if err := mapstructure.Decode(s.Metadata, &md); err != nil {
		return nil, err
	}

	return buildplan.BuildPlan{Dependency: buildplan.Dependency{Metadata: md}}, nil
}

// NewSpringBoot creates a new SpringBoot instance.  OK is true if the build plan contains a "jvm-application"
// dependency and a "Spring-Boot-Version" manifest key.
func NewSpringBoot(build build.Build) (SpringBoot, bool, error) {
	_, ok := build.BuildPlan[jvmapplication.Dependency]
	if !ok {
		return SpringBoot{}, false, nil
	}

	md, ok, err := NewMetadata(build.Application, build.Logger)
	if err != nil {
		return SpringBoot{}, false, err
	}

	if !ok {
		return SpringBoot{}, false, nil
	}

	return SpringBoot{
		md,
		build.Layers.Layer(Dependency),
		build.Layers,
	}, true, nil
}
