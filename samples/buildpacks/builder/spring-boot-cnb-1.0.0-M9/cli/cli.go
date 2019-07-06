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

package cli

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

// Dependency indicates that an application qualifies to have the Spring Boot CLI run its .groovy files.
const Dependency = "spring-boot-cli"

// CLI represents a Spring Boot CLI application.
type CLI struct {
	layer layers.DependencyLayer
}

// Contribute makes the contribution to launch.
func (c CLI) Contribute() error {
	return c.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		return helper.ExtractTarGz(artifact, layer.Root, 1)
	}, layers.Launch)
}

// NewCLI creates a new CLI instance.
func NewCLI(build build.Build) (CLI, error) {
	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return CLI{}, err
	}

	dep, err := deps.Best(Dependency, "", build.Stack)
	if err != nil {
		return CLI{}, err
	}

	return CLI{build.Layers.DependencyLayer(dep)}, nil
}
