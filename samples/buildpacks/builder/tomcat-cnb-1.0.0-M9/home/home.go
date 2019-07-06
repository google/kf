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

package home

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/tomcat-cnb/internal"
)

// TomcatDependency indicates that Tomcat is required for the web application.
const TomcatDependency = "tomcat"

type Home struct {
	layer  layers.DependencyLayer
	layers layers.Layers
}

func (h Home) Contribute() error {
	if err := h.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Extracting to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		return layer.OverrideLaunchEnv("CATALINA_HOME", layer.Root)
	}, layers.Launch); err != nil {
		return err
	}

	command := "catalina.sh run"

	return h.layers.WriteApplicationMetadata(layers.Metadata{
		Processes: layers.Processes{
			{"task", command},
			{"tomcat", command},
			{"web", command},
		},
	})
}

// NewHome creates a new CATALINA_HOME instance.
func NewHome(build build.Build) (Home, error) {
	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return Home{}, err
	}

	version, err := internal.Version(TomcatDependency, build.BuildPlan[TomcatDependency], build.Buildpack)
	if err != nil {
		return Home{}, err
	}

	dep, err := deps.Best(TomcatDependency, version, build.Stack)
	if err != nil {
		return Home{}, err
	}

	return Home{
		build.Layers.DependencyLayer(dep),
		build.Layers,
	}, nil
}
